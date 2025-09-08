package pager

import (
    "context"
    "fmt"
    "reflect"
    "strings"

    pagerpb "github.com/sky1core/proto-bun-page/proto/pager/v1"
    "github.com/uptrace/bun"
)


// ApplyAndScan applies pagination according to protobuf Page and scans results into dest.
// Flow:
//  1) Validate selector presence (page/cursor) and destination type
//  2) Infer model info and build order plan (with PK tiebreaker)
//  3) Normalize limit (default/clamp)
//  4) Decide mode and apply WHERE (cursor) or OFFSET (page)
//  5) Apply ORDER and LIMIT(+1), execute and trim
//  6) Build next cursor (cursor mode)
func (p *Pager) ApplyAndScan(ctx context.Context, q *bun.SelectQuery, in *pagerpb.Page, dest interface{}) (*pagerpb.Page, error) {
    if in == nil {
        in = &pagerpb.Page{}
    }

    hasCursor, hasPage := detectPresence(in)
    pageVal, cursorVal := getPageAndCursor(in)
    // Validate mutual exclusivity
    if hasPage && hasCursor {
        return nil, NewInvalidRequestError("cannot specify both page and cursor")
    }

    // Infer model info from destination
    if dest == nil {
        return nil, NewInvalidRequestError("destination must be a non-nil pointer to slice")
    }
    destType := reflect.TypeOf(dest)
    if destType.Kind() != reflect.Ptr {
        return nil, NewInvalidRequestError("destination must be a non-nil pointer to slice")
    }
    destType = destType.Elem()
    if destType.Kind() != reflect.Slice {
        return nil, NewInvalidRequestError("destination must be a non-nil pointer to slice")
    }
    modelType := destType.Elem()
    var model interface{}
    if modelType.Kind() == reflect.Ptr {
        model = reflect.New(modelType.Elem()).Interface()
    } else {
        model = reflect.New(modelType).Interface()
    }
    modelInfo, err := InferModelInfo(model)
    if err != nil {
        return nil, NewInternalError(fmt.Sprintf("failed to infer model info: %v", err))
    }

    // Build order specs from proto
    specs := make([]OrderSpec, 0, len(in.Order))
    for _, o := range in.Order {
        if o == nil || strings.TrimSpace(o.Key) == "" { continue }
        specs = append(specs, OrderSpec{Key: o.Key, Asc: o.Asc})
    }
    if len(specs) == 0 && len(p.opts.DefaultOrderSpecs) > 0 {
        specs = append(specs, p.opts.DefaultOrderSpecs...)
    }
    orderPlan, err := BuildOrderPlanFromSpecs(specs, modelInfo, p.opts.AllowedOrderKeys)
    if err != nil {
        if pe, ok := err.(*PagerError); ok {
            return nil, pe
        }
        return nil, NewInternalError(fmt.Sprintf("failed to build order plan: %v", err))
    }

    // Limit handling
    limit, clamped := normalizeLimit(in.Limit, p.opts)
    if clamped { p.logger.Warn("limit clamped from %d to %d (max)", in.Limit, p.opts.MaxLimit) }

    // Determine mode and apply WHERE
    mode := "offset"
    if hasCursor {
        mode = "cursor"
        // Empty cursor string means "from the beginning"; DecodeCursor returns nil
        cd, err := DecodeCursor(cursorVal, modelInfo)
        if err != nil {
            return nil, NewInvalidRequestError(fmt.Sprintf("invalid cursor: %v", err))
        }
        if cd != nil && len(cd.Values) > 0 {
            // Fetch anchor by single PK
            anchor := reflect.New(reflect.Indirect(reflect.ValueOf(model)).Type()).Interface()
            pkCol := "id"
            if len(modelInfo.PKColumns) > 0 { pkCol = modelInfo.PKColumns[0] }
            v, ok := cd.Values[pkCol]
            if !ok { return nil, NewInvalidRequestError("invalid cursor: missing pk") }
            // Normalize pk value to the model field type when possible
            if idx, ok := modelInfo.FieldIndexByColumn[pkCol]; ok {
                // Determine expected kind from model field
                mt := reflect.Indirect(reflect.ValueOf(model)).Type().Field(idx).Type
                v = coerceToKind(v, mt.Kind())
            }
            aq := q.DB().NewSelect().Model(anchor).Where(pkCol+" = ?", v).Limit(1)
            if err := aq.Scan(ctx); err != nil {
                return nil, NewStaleCursorError()
            }
            anchorVals, err := ExtractRowValues(anchor, orderPlan, modelInfo)
            if err != nil {
                return nil, NewInternalError(fmt.Sprintf("failed to extract anchor values: %v", err))
            }
            where, args2, err := BuildCursorWhere(&CursorData{Values: anchorVals}, orderPlan)
            if err != nil {
                return nil, NewInternalError(fmt.Sprintf("failed to build cursor where: %v", err))
            }
            if where != "" {
                q = q.Where(where, args2...)
            }
        }
    } else if hasPage {
        if pageVal < 1 {
            return nil, NewInvalidRequestError("page must be >= 1")
        }
        if pageVal > 1 {
            offset := (int(pageVal) - 1) * limit
            q = q.Offset(offset)
        }
    } else {
        // Neither page nor cursor specified: default to cursor mode
        mode = "cursor"
    }

    // Apply order and limit(+1)
    q = ApplyOrderToQuery(q, orderPlan)
    q = q.Limit(limit + 1)

    // Execute
    if err := q.Scan(ctx, dest); err != nil {
        return nil, NewInternalError(fmt.Sprintf("query execution failed: %v", err))
    }

    // Trim and build next cursor
    destValue := reflect.ValueOf(dest).Elem()
    rowCount := destValue.Len()
    hasMore := false
    if rowCount > limit {
        hasMore = true
        destValue.Set(destValue.Slice(0, limit))
        rowCount = limit
    }

    out := &pagerpb.Page{Limit: uint32(limit), Order: in.Order}
    
    if mode == "cursor" {
        if rowCount > 0 && hasMore {
            // Generate next cursor
            lastRow := destValue.Index(rowCount - 1)
            if lastRow.Kind() == reflect.Ptr { lastRow = lastRow.Elem() }
            values := make(map[string]interface{})
            for _, item := range orderPlan.Items {
                if idx, ok := modelInfo.FieldIndexByColumn[item.Column]; ok {
                    values[item.Column] = lastRow.Field(idx).Interface()
                }
            }
            if next, err := EncodeCursor(orderPlan, values, modelInfo); err == nil {
                out.Selector = &pagerpb.Page_Cursor{Cursor: next}
            }
        } else {
            // No next cursor
            out.Selector = &pagerpb.Page_Cursor{Cursor: ""}
        }
    } else {
        // Page mode - echo back the page number
        out.Selector = &pagerpb.Page_Page{Page: pageVal}
    }

    return out, nil
}

// detectPresence determines whether page/cursor selectors were explicitly provided.
func detectPresence(in *pagerpb.Page) (hasCursor, hasPage bool) {
    if in == nil { return false, false }
    
    // Check oneof selector field
    switch in.Selector.(type) {
    case *pagerpb.Page_Page:
        return false, true
    case *pagerpb.Page_Cursor:
        return true, false
    default:
        return false, false
    }
}

// normalizeLimit resolves the effective limit and whether it was clamped by MaxLimit.
func normalizeLimit(req uint32, opts *Options) (limit int, clamped bool) {
    limit = opts.DefaultLimit
    if req > 0 {
        limit = int(req)
        if opts.MaxLimit > 0 && limit > opts.MaxLimit {
            limit = opts.MaxLimit
            return limit, true
        }
    }
    return limit, false
}

// getPageAndCursor extracts values from oneof selector.
func getPageAndCursor(in *pagerpb.Page) (page uint32, cursor string) {
    if in == nil { return 0, "" }
    
    switch s := in.Selector.(type) {
    case *pagerpb.Page_Page:
        return s.Page, ""
    case *pagerpb.Page_Cursor:
        return 0, s.Cursor
    default:
        return 0, ""
    }
}
