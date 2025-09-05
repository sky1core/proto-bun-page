package pager

import (
    "fmt"
    "reflect"
    "strings"
)

type ModelInfo struct {
    TableName    string
    PKColumns    []string
    ColumnToKey  map[string]string
    KeyToColumn  map[string]string
}

func InferModelInfo(model interface{}) (*ModelInfo, error) {
	info := &ModelInfo{
		ColumnToKey: make(map[string]string),
		KeyToColumn: make(map[string]string),
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        bunTag := field.Tag.Get("bun")
        if bunTag == "" {
            continue
        }

        parts := strings.Split(bunTag, ",")
        columnName := parts[0]
        if columnName == "" {
            // No implicit snake_case fallback: column must be specified in bun tag
            continue
        }

        // Logical key equals bun column name
        info.ColumnToKey[columnName] = columnName
        info.KeyToColumn[columnName] = columnName

        for _, part := range parts {
            if part == "pk" {
                info.PKColumns = append(info.PKColumns, columnName)
            }
        }
    }

    if len(info.PKColumns) == 0 {
        info.PKColumns = []string{"id"}
        info.ColumnToKey["id"] = "id"
        info.KeyToColumn["id"] = "id"
    }
    if len(info.PKColumns) > 1 {
        return nil, fmt.Errorf("composite primary key is not supported")
    }

	return info, nil
}
