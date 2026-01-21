package commondb

import "strings"

// SplitWhereHaving splits SQL conditions into WHERE (regular fields) and HAVING (aggregates)
func SplitWhereHaving(sqlCondition string) (whereClause, havingClause string) {
	sqlCondition = strings.TrimSpace(sqlCondition)
	if strings.HasPrefix(sqlCondition, "(") && strings.HasSuffix(sqlCondition, ")") {
		sqlCondition = sqlCondition[1 : len(sqlCondition)-1]
	}

	var whereParts, havingParts []string
	for _, part := range splitOnAND(sqlCondition) {
		if part = strings.TrimSpace(part); part == "" {
			continue
		}
		if strings.Contains(part, "COUNT(") {
			havingParts = append(havingParts, part)
		} else {
			whereParts = append(whereParts, part)
		}
	}

	if len(whereParts) > 0 {
		whereClause = strings.Join(whereParts, " AND ")
	}
	if len(havingParts) > 0 {
		havingClause = strings.Join(havingParts, " AND ")
	}
	return
}

func splitOnAND(condition string) []string {
	var parts []string
	var current strings.Builder
	parenDepth := 0

	for i := 0; i < len(condition); i++ {
		switch condition[i] {
		case '(':
			parenDepth++
			current.WriteByte(condition[i])
		case ')':
			parenDepth--
			current.WriteByte(condition[i])
		default:
			if parenDepth == 0 && i+5 <= len(condition) && condition[i:i+5] == " AND " {
				parts = append(parts, current.String())
				current.Reset()
				i += 4
			} else {
				current.WriteByte(condition[i])
			}
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
