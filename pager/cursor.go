package pager

import (
    "encoding/base64"
    "fmt"
    "math"
    "reflect"
    "strconv"
)

// CursorData represents decoded cursor values carried by a cursor token.
type CursorData struct {
    Values map[string]interface{}
}

// EncodeCursor creates a cursor string from row values
func EncodeCursor(orderPlan *OrderPlan, row map[string]interface{}, modelInfo *ModelInfo) (string, error) {
    cursorData := CursorData{Values: make(map[string]interface{})}

    // Encode single primary key value for cursor token
    pk := firstPKColumn(modelInfo)
    if val, ok := row[pk]; ok {
        cursorData.Values[pk] = val
    }

    // Build value-only payload in PK order: v1,v2,...
    // Just single value string
    s := ""
    if v, ok := cursorData.Values[pk]; ok {
        s = fmt.Sprint(v)
    }
    cursor := base64.URLEncoding.EncodeToString([]byte(s))
    return cursor, nil
}

// DecodeCursor decodes a cursor string into values
func DecodeCursor(cursor string, modelInfo *ModelInfo) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}

    decoded, err := base64.URLEncoding.DecodeString(cursor)
    if err != nil {
        return nil, NewInvalidRequestError("invalid cursor format")
    }
    s := string(decoded)
    cd := &CursorData{Values: map[string]interface{}{}}
    if len(s) == 0 {
        return cd, nil
    }
    pk := firstPKColumn(modelInfo)
    sv := s
    if iv, err := strconv.ParseInt(sv, 10, 64); err == nil {
        cd.Values[pk] = iv
    } else {
        cd.Values[pk] = sv
    }
    return cd, nil
}

// ExtractRowValues extracts values from a row for cursor creation
func ExtractRowValues(row interface{}, orderPlan *OrderPlan, modelInfo *ModelInfo) (map[string]interface{}, error) {
    values := make(map[string]interface{})

    v := reflect.ValueOf(row)
    if v.Kind() == reflect.Ptr { v = v.Elem() }
    if v.Kind() != reflect.Struct {
        return nil, fmt.Errorf("row must be a struct or pointer to struct")
    }

    for _, item := range orderPlan.Items {
        if idx, ok := modelInfo.FieldIndexByColumn[item.Column]; ok {
            values[item.Column] = v.Field(idx).Interface()
        }
    }
    return values, nil
}

// coerceToKind converts v into a value assignable for the given reflect.Kind when reasonable.
// For unsupported combinations, returns the original v.
func coerceToKind(v interface{}, k reflect.Kind) interface{} {
    switch k {
    case reflect.Int, reflect.Int32, reflect.Int64:
        switch nv := v.(type) {
        case int64:
            return nv
        case int:
            return int64(nv)
        case int32:
            return int64(nv)
        case uint32:
            return int64(nv)
        case uint64:
            if nv > uint64(math.MaxInt64) { return v }
            return int64(nv)
        case float64:
            return int64(nv)
        case string:
            if iv, err := strconv.ParseInt(nv, 10, 64); err == nil { return iv }
            return v
        }
    case reflect.Uint, reflect.Uint32, reflect.Uint64:
        switch nv := v.(type) {
        case int64:
            if nv < 0 { return v }
            return uint64(nv)
        case int:
            if nv < 0 { return v }
            return uint64(nv)
        case int32:
            if nv < 0 { return v }
            return uint64(nv)
        case uint32:
            return uint64(nv)
        case uint64:
            return nv
        case float64:
            if nv < 0 { return v }
            return uint64(nv)
        case string:
            if uv, err := strconv.ParseUint(nv, 10, 64); err == nil { return uv }
            return v
        }
    case reflect.String:
        switch nv := v.(type) {
        case string:
            return nv
        default:
            return fmt.Sprint(nv)
        }
    }
    return v
}
