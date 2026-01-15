package rsql

// FieldMapping maps public API field names to database column names
// This prevents exposing internal database structure and provides
// strict whitelisting for allowed filter fields
type FieldMapping struct {
	APIName    string
	DBColumn   string
	IsVirtual  bool // True for computed fields like 'likes'
	RequiresAuth bool // True for fields that require authentication to filter
}

// EntryFieldMappings defines all filterable fields for Entry model
// Only fields in this map can be used in filters
var EntryFieldMappings = map[string]FieldMapping{
	// Basic fields (always accessible)
	"id":       {APIName: "id", DBColumn: "id", IsVirtual: false, RequiresAuth: false},
	"date":     {APIName: "date", DBColumn: "date", IsVirtual: false, RequiresAuth: false},
	"time":     {APIName: "time", DBColumn: "time", IsVirtual: false, RequiresAuth: false},
	"datetime": {APIName: "datetime", DBColumn: "datetime", IsVirtual: false, RequiresAuth: false},
	"status":   {APIName: "status", DBColumn: "status", IsVirtual: false, RequiresAuth: false},
	"cl":       {APIName: "cl", DBColumn: "cl", IsVirtual: false, RequiresAuth: false},
	"place":    {APIName: "place", DBColumn: "place", IsVirtual: false, RequiresAuth: false},
	"olsug":    {APIName: "olsug", DBColumn: "olsug", IsVirtual: false, RequiresAuth: false},
	"enheter":  {APIName: "enheter", DBColumn: "enheter", IsVirtual: false, RequiresAuth: false},
	"report":   {APIName: "report", DBColumn: "report", IsVirtual: false, RequiresAuth: false},
	
	// Potentially sensitive fields (require viewing permission to filter)
	"msg":   {APIName: "msg", DBColumn: "msg", IsVirtual: false, RequiresAuth: true},
	"sig":   {APIName: "sig", DBColumn: "sig", IsVirtual: false, RequiresAuth: true},
	"email": {APIName: "email", DBColumn: "email", IsVirtual: false, RequiresAuth: true},
	
	// Virtual/computed fields
	"likes":           {APIName: "likes", DBColumn: "", IsVirtual: true, RequiresAuth: false},
	"secret":          {APIName: "secret", DBColumn: "", IsVirtual: true, RequiresAuth: false},
	"personal_secret": {APIName: "personal_secret", DBColumn: "", IsVirtual: true, RequiresAuth: false},
}

// GetFieldMapping returns the field mapping for an API field name
// Returns nil if field is not whitelisted
func GetFieldMapping(apiFieldName string) *FieldMapping {
	if mapping, exists := EntryFieldMappings[apiFieldName]; exists {
		return &mapping
	}
	return nil
}

// IsFieldAllowed checks if a field can be filtered by the given viewer
// Returns true if field is allowed, false if filtering should be rejected
func IsFieldAllowed(apiFieldName string, viewerMemberID *int64, canViewSecretContent bool) bool {
	mapping := GetFieldMapping(apiFieldName)
	if mapping == nil {
		return false // Field not in whitelist
	}
	
	// If field requires auth, viewer must be authenticated
	if mapping.RequiresAuth {
		if viewerMemberID == nil {
			return false // Not authenticated
		}
		if !canViewSecretContent {
			return false // Authenticated but can't view secret content
		}
	}
	
	return true
}
