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
func (p *Pager) ApplyAndScan(ctx context.Context, q *bun.SelectQuery, in *pagerpb.Page, dest interface{}) (*pagerpb.Page, error) {
    if in == nil {
        in = &pagerpb.Page{}
    }

    // Validate mutual exclusivity and page range
    if in.Page > 0 && strings.TrimSpace(in.Cursor) != "" {
        return nil, NewInvalidRequestError("cannot specify both page and cursor")
    }

    // Infer model info from destination
    destType := reflect.TypeOf(dest)
    if destType.Kind() != reflect.Ptr {
        return nil, NewInternalError("destination must be a pointer to slice")
    }
    destType = destType.Elem()
    if destType.Kind() != reflect.Slice {
        return nil, NewInternalError("destination must be a pointer to slice")
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
        specs = append(specs, OrderSpec{Key: o.Key, Desc: o.Desc})
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
    limit := p.opts.DefaultLimit
    if in.Limit > 0 {
        limit = int(in.Limit)
        if limit > p.opts.MaxLimit {
            p.logger.Warn("limit clamped from %d to %d (max)", limit, p.opts.MaxLimit)
            limit = p.opts.MaxLimit
        }
    }

    // Determine mode and apply WHERE
    mode := "offset"
    if strings.TrimSpace(in.Cursor) != "" {
        mode = "cursor"
        cd, err := DecodeCursor(in.Cursor, modelInfo)
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
            switch nv := v.(type) {
            case float64:
                v = int64(nv)
            case int:
                v = int64(nv)
            case int32:
                v = int64(nv)
            case uint64:
                v = int64(nv)
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
    } else if in.Page == 0 {
        // default to cursor mode when neither page nor cursor provided
        mode = "cursor"
    } else if in.Page > 1 {
        offset := (int(in.Page) - 1) * limit
        q = q.Offset(offset)
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

    out := &pagerpb.Page{Limit: uint32(limit), Order: in.Order, Page: in.Page, Cursor: ""}

    if hasMore || (mode == "cursor" && rowCount > 0) {
        if mode == "cursor" && rowCount > 0 {
            lastRow := destValue.Index(rowCount - 1)
            if lastRow.Kind() == reflect.Ptr { lastRow = lastRow.Elem() }
            values := make(map[string]interface{})
            for _, item := range orderPlan.Items {
                if idx, ok := modelInfo.FieldIndexByColumn[item.Column]; ok {
                    values[item.Column] = lastRow.Field(idx).Interface()
                }
            }
            if next, err := EncodeCursor(orderPlan, values, modelInfo); err == nil {
                out.Cursor = next
            }
        }
    }

    return out, nil
}
