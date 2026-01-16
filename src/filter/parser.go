package filter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// FilterParser wraps RSQL parsing with GORM integration and security
type FilterParser struct {
	schema     *Schema
	userScopes []string
}

// NewFilterParser creates a new parser for a schema
func NewFilterParser(schema *Schema, userScopes []string) *FilterParser {
	return &FilterParser{
		schema:     schema,
		userScopes: userScopes,
	}
}

// ParseAndApply parses the RSQL query string and applies it to GORM db instance
// Supports FIQL/RSQL syntax: field==value, field!=value, field>value, etc.
// Logical operators: ; (AND), , (OR)
func (p *FilterParser) ParseAndApply(db *gorm.DB, queryString string) (*gorm.DB, error) {
	if queryString == "" {
		return db, nil
	}

	// Parse the RSQL string into filter tokens
	// For now, we support AND (;) but not OR (,) for simplicity
	// Split on AND operator first
	andClauses := strings.Split(queryString, ";")

	for _, clause := range andClauses {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}

		// Parse individual filter clause
		var err error
		db, err = p.parseClause(db, clause)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// parseClause parses a single filter clause like "field==value" or "field>10"
func (p *FilterParser) parseClause(db *gorm.DB, clause string) (*gorm.DB, error) {
	// RSQL operators in order of specificity (longer first to match correctly)
	operators := []struct {
		rsql string
		sql  string
	}{
		{"==", "="},
		{"!=", "!="},
		{"=ge=", ">="},
		{"=gt=", ">"},
		{"=le=", "<="},
		{"=lt=", "<"},
		{"=like=", "LIKE"},
		{"=in=", "IN"},
		{"=out=", "NOT IN"},
	}

	// Try to match an operator
	for _, op := range operators {
		if strings.Contains(clause, op.rsql) {
			parts := strings.SplitN(clause, op.rsql, 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid filter syntax: %s", clause)
			}

			field := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes from value if present
			value = strings.Trim(value, "\"'")

			// Validate the field is allowed
			fieldMapping, err := p.schema.ValidateField(field, p.userScopes)
			if err != nil {
				return nil, fmt.Errorf("invalid filter field: %w", err)
			}

			// Apply the filter based on operator
			return p.applyFilter(db, fieldMapping.DBColumn, op.sql, value, op.rsql)
		}
	}

	return nil, fmt.Errorf("no valid operator found in clause: %s", clause)
}

// applyFilter applies a single filter condition to the GORM query
// Uses parameterized queries to prevent SQL injection
func (p *FilterParser) applyFilter(db *gorm.DB, column string, sqlOp string, value string, rsqlOp string) (*gorm.DB, error) {
	switch sqlOp {
	case "=", "!=", ">", ">=", "<", "<=":
		return db.Where(column+" "+sqlOp+" ?", value), nil

	case "LIKE":
		// Support wildcards: if no % present, add them
		if !strings.Contains(value, "%") {
			value = "%" + value + "%"
		}
		return db.Where(column+" LIKE ?", value), nil

	case "IN", "NOT IN":
		// Parse comma-separated values for IN operator
		values := strings.Split(value, ",")
		cleanValues := make([]string, len(values))
		for i, v := range values {
			cleanValues[i] = strings.TrimSpace(v)
		}
		return db.Where(column+" "+sqlOp+" ?", cleanValues), nil

	default:
		return nil, fmt.Errorf("unsupported operator: %s", sqlOp)
	}
}

// ParseSorting parses and applies sorting from the sort string
// Format: "field1,-field2" means field1 ASC, field2 DESC
func (p *FilterParser) ParseSorting(db *gorm.DB, sortString string) (*gorm.DB, error) {
	if sortString == "" {
		return db, nil
	}

	// Parse sort fields: "field1,-field2" means field1 ASC, field2 DESC
	sortFields := strings.Split(sortString, ",")

	for _, sortField := range sortFields {
		sortField = strings.TrimSpace(sortField)
		if sortField == "" {
			continue
		}

		// Check for descending order (prefix with -)
		direction := "ASC"
		fieldName := sortField
		if strings.HasPrefix(sortField, "-") {
			direction = "DESC"
			fieldName = sortField[1:]
		}

		// Validate and translate field
		fieldMapping, err := p.schema.ValidateField(fieldName, p.userScopes)
		if err != nil {
			return nil, fmt.Errorf("invalid sort field: %w", err)
		}

		// Prevent SQL injection by validating direction
		if direction != "ASC" && direction != "DESC" {
			return nil, fmt.Errorf("invalid sort direction: %s", direction)
		}

		// Apply sorting (safe because direction is validated)
		db = db.Order(fieldMapping.DBColumn + " " + direction)
	}

	return db, nil
}
