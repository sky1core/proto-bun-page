package pager

import (
    "encoding/base64"
    "fmt"
    "reflect"
    "strconv"
)

// CursorData represents the decoded cursor values
type CursorData struct {
    Values map[string]interface{}
    Mode   string // "offset" or "cursor"
}

// EncodeCursor creates a cursor string from row values
func EncodeCursor(orderPlan *OrderPlan, row map[string]interface{}, modelInfo *ModelInfo) (string, error) {
    cursorData := CursorData{
        Values: make(map[string]interface{}),
        Mode:   "cursor",
    }

    // Encode single primary key value for cursor token
    pk := "id"
    if len(modelInfo.PKColumns) > 0 { pk = modelInfo.PKColumns[0] }
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
        return nil, fmt.Errorf("invalid cursor format: %w", err)
    }
    s := string(decoded)
    cd := &CursorData{Values: map[string]interface{}{}, Mode: "cursor"}
    if len(s) == 0 {
        return cd, nil
    }
    pk := "id"
    if len(modelInfo.PKColumns) > 0 { pk = modelInfo.PKColumns[0] }
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

func indexOfComma(s string) int {
    for i := 0; i < len(s); i++ {
        if s[i] == ',' {
            return i
        }
    }
    return -1
}
