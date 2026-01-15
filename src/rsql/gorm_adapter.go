package rsql

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
	
	"github.com/rbicker/go-rsql"
	"gorm.io/gorm"
)

// GormAdapter creates GORM-compatible WHERE clauses from RSQL
type GormAdapter struct {
	parser *rsql.Parser
	ctx    FilterContext
}

// NewGormAdapter creates a new GORM adapter with field mappings and permissions
func NewGormAdapter(ctx FilterContext) (*GormAdapter, error) {
	// Define GORM operators
	operators := []rsql.Operator{
		{
			Operator: "==",
			Formatter: func(key, value string) string {
				// Check field access permission
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, "=", value)
				}
				
				return fmt.Sprintf("%s = '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "!=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, "!=", value)
				}
				
				return fmt.Sprintf("%s != '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "=gt=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, ">", value)
				}
				
				return fmt.Sprintf("%s > '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "=ge=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, ">=", value)
				}
				
				return fmt.Sprintf("%s >= '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "=lt=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, "<", value)
				}
				
				return fmt.Sprintf("%s < '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "=le=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return formatVirtualField(key, "<=", value)
				}
				
				return fmt.Sprintf("%s <= '%s'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
		{
			Operator: "=in=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return fmt.Sprintf(`{"error":"operator 'in' not supported for virtual field '%s'}`, key)
				}
				
				// Parse list: (val1,val2,val3) -> 'val1','val2','val3'
				values := parseListValue(value)
				quotedValues := make([]string, len(values))
				for i, v := range values {
					quotedValues[i] = "'" + escapeSQLValue(v) + "'"
				}
				
				return fmt.Sprintf("%s IN (%s)", mapping.DBColumn, strings.Join(quotedValues, ","))
			},
		},
		{
			Operator: "=out=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return fmt.Sprintf(`{"error":"operator 'out' not supported for virtual field '%s'}`, key)
				}
				
				// Parse list: (val1,val2,val3) -> 'val1','val2','val3'
				values := parseListValue(value)
				quotedValues := make([]string, len(values))
				for i, v := range values {
					quotedValues[i] = "'" + escapeSQLValue(v) + "'"
				}
				
				return fmt.Sprintf("%s NOT IN (%s)", mapping.DBColumn, strings.Join(quotedValues, ","))
			},
		},
		{
			Operator: "=like=",
			Formatter: func(key, value string) string {
				if !IsFieldAllowed(key, ctx.ViewerMemberID, ctx.CanViewSecretContent) {
					return fmt.Sprintf(`{"error":"field '%s' not allowed"}`, key)
				}
				
				mapping := GetFieldMapping(key)
				if mapping == nil {
					return fmt.Sprintf(`{"error":"field '%s' not whitelisted"}`, key)
				}
				
				if mapping.IsVirtual {
					return fmt.Sprintf(`{"error":"operator 'like' not supported for virtual field '%s'}`, key)
				}
				
				return fmt.Sprintf("%s LIKE '%%%s%%'", mapping.DBColumn, escapeSQLValue(value))
			},
		},
	}
	
	// Create parser with GORM operators and SQL formatters
	parser, err := rsql.NewParser(
		rsql.WithOperators(operators...),
		withSQLFormatters(), // Custom option for SQL AND/OR formatters
	)
	if err != nil {
		return nil, err
	}
	
	return &GormAdapter{
		parser: parser,
		ctx:    ctx,
	}, nil
}

// withSQLFormatters creates a parser option that sets SQL-style AND/OR formatters
// This mirrors the Mongo() function from the library but for SQL
func withSQLFormatters() func(*rsql.Parser) error {
	return func(parser *rsql.Parser) error {
		// Use reflection to set the private andFormatter and orFormatter fields
		// This is necessary because the library doesn't expose setter functions
		parserValue := reflect.ValueOf(parser).Elem()
		
		// Set AND formatter
		andFormatterField := parserValue.FieldByName("andFormatter")
		if andFormatterField.IsValid() {
			andFormatter := func(ss []string) string {
				if len(ss) == 0 {
					return ""
				}
				if len(ss) == 1 {
					return ss[0]
				}
				return "(" + strings.Join(ss, " AND ") + ")"
			}
			
			// Make the field settable
			andFormatterField = reflect.NewAt(andFormatterField.Type(), unsafe.Pointer(andFormatterField.UnsafeAddr())).Elem()
			andFormatterField.Set(reflect.ValueOf(andFormatter))
		}
		
		// Set OR formatter
		orFormatterField := parserValue.FieldByName("orFormatter")
		if orFormatterField.IsValid() {
			orFormatter := func(ss []string) string {
				if len(ss) == 0 {
					return ""
				}
				if len(ss) == 1 {
					return ss[0]
				}
				return "(" + strings.Join(ss, " OR ") + ")"
			}
			
			// Make the field settable
			orFormatterField = reflect.NewAt(orFormatterField.Type(), unsafe.Pointer(orFormatterField.UnsafeAddr())).Elem()
			orFormatterField.Set(reflect.ValueOf(orFormatter))
		}
		
		return nil
	}
}

// Process parses RSQL query and applies it to GORM query
func (g *GormAdapter) Process(db *gorm.DB, rsqlQuery string) (*gorm.DB, error) {
	if rsqlQuery == "" {
		return db, nil
	}
	
	// Parse RSQL to SQL WHERE clause
	whereClause, err := g.parser.Process(rsqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSQL: %w", err)
	}
	
	// Check for error markers in the generated clause
	if strings.Contains(whereClause, `{"error":`) {
		// Extract error message
		start := strings.Index(whereClause, `"error":"`) + 9
		end := strings.Index(whereClause[start:], `"`)
		if end > 0 {
			errMsg := whereClause[start : start+end]
			return nil, fmt.Errorf("%s", errMsg)
		}
		return nil, fmt.Errorf("invalid filter query")
	}
	
	// Apply WHERE clause to GORM query
	return db.Where(whereClause), nil
}

// formatVirtualField creates SQL for virtual/computed fields
func formatVirtualField(field, operator, value string) string {
	switch field {
	case "likes":
		subQuery := "(SELECT COUNT(*) FROM 2003_likes WHERE 2003_likes.id = cl2003_msgs.id)"
		return fmt.Sprintf("%s %s '%s'", subQuery, operator, escapeSQLValue(value))
		
	case "secret":
		existsQuery := "EXISTS (SELECT 1 FROM cl2003_permissions WHERE cl2003_permissions.id = cl2003_msgs.id)"
		if value == "true" || value == "1" {
			if operator == "=" || operator == "==" {
				return existsQuery
			} else {
				return "NOT " + existsQuery
			}
		} else {
			if operator == "=" || operator == "==" {
				return "NOT " + existsQuery
			} else {
				return existsQuery
			}
		}
		
	case "personal_secret":
		existsQuery := "EXISTS (SELECT 1 FROM cl2003_permissions WHERE cl2003_permissions.id = cl2003_msgs.id AND cl2003_permissions.user_id != 0)"
		if value == "true" || value == "1" {
			if operator == "=" || operator == "==" {
				return existsQuery
			} else {
				return "NOT " + existsQuery
			}
		} else {
			if operator == "=" || operator == "==" {
				return "NOT " + existsQuery
			} else {
				return existsQuery
			}
		}
		
	default:
		return fmt.Sprintf(`{"error":"virtual field '%s' not implemented"}`, field)
	}
}

// escapeSQLValue escapes single quotes in SQL values
func escapeSQLValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// parseListValue parses RSQL list format: (val1,val2,val3) -> []string{"val1","val2","val3"}
func parseListValue(value string) []string {
	// Remove parentheses
	value = strings.TrimPrefix(value, "(")
	value = strings.TrimSuffix(value, ")")
	
	// Remove quotes from each value and split
	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "'\"")
		result[i] = part
	}
	
	return result
}
