package pager

import (
	"fmt"
	"strings"

	"github.com/uptrace/bun"
)

type OrderItem struct {
	Column    string
	Direction string
}

type OrderPlan struct {
    Items []OrderItem
}

type OrderSpec struct {
    Key  string
    Desc bool
}

// BuildOrderPlanFromSpecs builds an OrderPlan from structured specs (preferred path).
func BuildOrderPlanFromSpecs(specs []OrderSpec, modelInfo *ModelInfo, allowedKeys []string) (*OrderPlan, error) {
    plan := &OrderPlan{}

    allowSet := map[string]struct{}{}
    if len(allowedKeys) > 0 {
        for _, k := range allowedKeys {
            k = strings.TrimSpace(k)
            if k == "" { continue }
            allowSet[k] = struct{}{}
        }
    }

    // track to dedupe by column while preserving last occurrence order
    for _, s := range specs {
        nk := strings.TrimSpace(s.Key)
        if nk == "" { continue }
        if len(allowSet) > 0 {
            if _, ok := allowSet[nk]; !ok {
                return nil, NewInvalidRequestError("unsupported order key: " + nk)
            }
        }
        column, exists := modelInfo.KeyToColumn[nk]
        if !exists {
            return nil, NewInvalidRequestError("unsupported order key: " + nk)
        }
        dir := "ASC"
        if s.Desc { dir = "DESC" }
        // remove previous occurrence of this column, if any
        if len(plan.Items) > 0 {
            out := plan.Items[:0]
            for _, it := range plan.Items {
                if it.Column != column { out = append(out, it) }
            }
            plan.Items = out
        }
        plan.Items = append(plan.Items, OrderItem{Column: column, Direction: dir})
    }

    // Resolve PK column
    pkCol := "id"
    if len(modelInfo.PKColumns) > 0 {
        pkCol = modelInfo.PKColumns[0]
    }
    // Ensure PK tiebreaker present following last effective direction
    present := map[string]struct{}{}
    for _, it := range plan.Items { present[it.Column] = struct{}{} }
    lastDir := "DESC"
    if len(plan.Items) > 0 { lastDir = plan.Items[len(plan.Items)-1].Direction }
    if _, ok := present[pkCol]; !ok {
        plan.Items = append(plan.Items, OrderItem{Column: pkCol, Direction: lastDir})
    }
    if len(plan.Items) == 0 {
        plan.Items = append(plan.Items, OrderItem{Column: pkCol, Direction: "DESC"})
    }
    return plan, nil
}

// Order plans are constructed from structured specs via BuildOrderPlanFromSpecs.

func BuildCursorWhere(cursorData *CursorData, orderPlan *OrderPlan) (string, []interface{}, error) {
	if cursorData == nil || len(cursorData.Values) == 0 {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	
	// Build OR-chain WHERE clause for cursor pagination
	// Example for (a DESC, b ASC, id ASC):
	// WHERE (a < ?) OR (a = ? AND b > ?) OR (a = ? AND b = ? AND id > ?)
	
	for i := 0; i <= len(orderPlan.Items)-1; i++ {
		var condition []string
		
		// Build equality conditions for all columns before the current one
		for j := 0; j < i; j++ {
			item := orderPlan.Items[j]
			if val, ok := cursorData.Values[item.Column]; ok {
				condition = append(condition, fmt.Sprintf("%s = ?", item.Column))
				args = append(args, val)
			}
		}
		
		// Add the inequality condition for the current column
		if i < len(orderPlan.Items) {
			item := orderPlan.Items[i]
			if val, ok := cursorData.Values[item.Column]; ok {
				op := ">"
				if item.Direction == "DESC" {
					op = "<"
				}
				condition = append(condition, fmt.Sprintf("%s %s ?", item.Column, op))
				args = append(args, val)
			}
		}
		
		if len(condition) > 0 {
			conditions = append(conditions, "("+strings.Join(condition, " AND ")+")")
		}
	}
	
	if len(conditions) == 0 {
		return "", nil, nil
	}
	
	whereClause := "(" + strings.Join(conditions, " OR ") + ")"
	return whereClause, args, nil
}

func ApplyOrderToQuery(q *bun.SelectQuery, orderPlan *OrderPlan) *bun.SelectQuery {
	for _, item := range orderPlan.Items {
		if item.Direction == "DESC" {
			q = q.Order(item.Column + " DESC")
		} else {
			q = q.Order(item.Column + " ASC")
		}
	}
	return q
}
