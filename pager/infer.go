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
			columnName = toSnakeCase(field.Name)
		}

		key := toSnakeCase(field.Name)
		info.ColumnToKey[columnName] = key
		info.KeyToColumn[key] = columnName

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

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
