package filter

import (
	"fmt"

	"github.com/sebastiw/sidan-backend/src/models"
)

// FieldMapping defines the public API field name to database column mapping
type FieldMapping struct {
	PublicName   string // API field name
	DBColumn     string // Database column name
	Type         string // Field type (string, int, date, etc.)
	Filterable   bool   // Can this field be used in filters?
	RequireScope string // Scope required to filter this field (optional)
}

// Schema defines allowed fields per resource
type Schema struct {
	Model    interface{}
	Fields   map[string]FieldMapping
	ACLField string // Field used for ACL checks (e.g., "id" for entries)
}

// EntrySchema defines the schema for entry filtering
var EntrySchema = Schema{
	Model:    models.Entry{},
	ACLField: "id",
	Fields: map[string]FieldMapping{
		"id":        {PublicName: "id", DBColumn: "id", Type: "int", Filterable: true},
		"signature": {PublicName: "signature", DBColumn: "sig", Type: "string", Filterable: true},
		"email":     {PublicName: "email", DBColumn: "email", Type: "string", Filterable: true},
		"place":     {PublicName: "place", DBColumn: "place", Type: "string", Filterable: true},
		"date":      {PublicName: "date", DBColumn: "datetime", Type: "datetime", Filterable: true},
		"status":    {PublicName: "status", DBColumn: "status", Type: "int", Filterable: true},
		"cl":        {PublicName: "cl", DBColumn: "cl", Type: "int", Filterable: true},
		// Note: "message" (msg) and "likes" are NOT included
		// - message: prevents side-channel attacks on "hemlis" content
		// - likes: stored in separate table (2003_likes), not directly filterable
	},
}

// ValidateField checks if a field is allowed for filtering based on user permissions
func (s *Schema) ValidateField(publicField string, userScopes []string) (FieldMapping, error) {
	field, exists := s.Fields[publicField]
	if !exists {
		return FieldMapping{}, fmt.Errorf("field '%s' does not exist or is not filterable", publicField)
	}

	if !field.Filterable {
		return FieldMapping{}, fmt.Errorf("field '%s' is not filterable", publicField)
	}

	// Check if scope is required
	if field.RequireScope != "" {
		hasScope := false
		for _, scope := range userScopes {
			if scope == field.RequireScope {
				hasScope = true
				break
			}
		}
		if !hasScope {
			return FieldMapping{}, fmt.Errorf("insufficient permissions to filter by '%s'", publicField)
		}
	}

	return field, nil
}

// TranslateField translates public field name to database column
func (s *Schema) TranslateField(publicField string) (string, error) {
	field, exists := s.Fields[publicField]
	if !exists {
		return "", fmt.Errorf("unknown field: %s", publicField)
	}
	return field.DBColumn, nil
}

// GetAllowedFields returns list of public field names the user can filter by
func (s *Schema) GetAllowedFields(userScopes []string) []string {
	var allowed []string
	for _, field := range s.Fields {
		if !field.Filterable {
			continue
		}
		// Check scope requirement
		if field.RequireScope != "" {
			hasScope := false
			for _, scope := range userScopes {
				if scope == field.RequireScope {
					hasScope = true
					break
				}
			}
			if !hasScope {
				continue
			}
		}
		allowed = append(allowed, field.PublicName)
	}
	return allowed
}
