# RSQL Implementation Plan for Sidan Backend

## Overview
This document outlines the implementation plan for advanced API filtering using RSQL/RQL with GORM, integrating with existing ACL system.

## Recommended Approach: Hybrid with libatomic/rql

### Phase 1: The Filtering Engine (Foundation)

#### Story 1.1: Schema Mapping & Whitelisting

**Implementation:**
```go
// src/filter/schema.go
package filter

import "github.com/sebastiw/sidan-backend/src/models"

// FieldMapping defines the public API field name to database column mapping
type FieldMapping struct {
    PublicName   string   // API field name
    DBColumn     string   // Database column name
    Type         string   // Field type (string, int, date, etc.)
    Filterable   bool     // Can this field be used in filters?
    RequireScope string   // Scope required to filter this field (optional)
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
        "message":   {PublicName: "message", DBColumn: "msg", Type: "string", Filterable: true},
        "signature": {PublicName: "signature", DBColumn: "sig", Type: "string", Filterable: true},
        "email":     {PublicName: "email", DBColumn: "email", Type: "string", Filterable: true},
        "place":     {PublicName: "place", DBColumn: "place", Type: "string", Filterable: true},
        "date":      {PublicName: "date", DBColumn: "datetime", Type: "datetime", Filterable: true},
        "likes":     {PublicName: "likes", DBColumn: "likes", Type: "int", Filterable: true},
        "secret":    {PublicName: "secret", DBColumn: "secret", Type: "bool", Filterable: false}, // Not filterable by users
    },
}

// ArticleSchema defines the schema for article filtering
var ArticleSchema = Schema{
    Model:    models.Article{},
    ACLField: "Id",
    Fields: map[string]FieldMapping{
        "id":     {PublicName: "id", DBColumn: "Id", Type: "int", Filterable: true, RequireScope: "read:article"},
        "header": {PublicName: "header", DBColumn: "header", Type: "string", Filterable: true, RequireScope: "read:article"},
        "body":   {PublicName: "body", DBColumn: "body", Type: "string", Filterable: true, RequireScope: "read:article"},
        "date":   {PublicName: "date", DBColumn: "datetime", Type: "datetime", Filterable: true, RequireScope: "read:article"},
    },
}

// ValidateField checks if a field is allowed for filtering based on user permissions
func (s *Schema) ValidateField(publicField string, userScopes []string) (FieldMapping, error) {
    field, exists := s.Fields[publicField]
    if !exists {
        return FieldMapping{}, fmt.Errorf("field '%s' does not exist", publicField)
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
```

#### Story 1.2: RSQL Parsing & Translation

**Implementation using libatomic/rql:**
```go
// src/filter/parser.go
package filter

import (
    "fmt"
    "gorm.io/gorm"
    "github.com/libatomic/rql"
)

// RQLParser wraps the rql parser with GORM integration
type RQLParser struct {
    schema *Schema
}

// NewRQLParser creates a new parser for a schema
func NewRQLParser(schema *Schema) *RQLParser {
    return &RQLParser{schema: schema}
}

// ParseAndApply parses the query string and applies it to GORM
func (p *RQLParser) ParseAndApply(db *gorm.DB, queryString string, userScopes []string) (*gorm.DB, error) {
    if queryString == "" {
        return db, nil
    }

    // Parse the query string
    params := rql.NewParser(
        // Use GORM's column naming convention
        rql.WithColumnFn(func(field string) string {
            // Validate field is allowed
            fieldMapping, err := p.schema.ValidateField(field, userScopes)
            if err != nil {
                return "" // Will cause error in execution
            }
            return fieldMapping.DBColumn
        }),
    )

    query, err := params.Parse(queryString)
    if err != nil {
        return nil, fmt.Errorf("failed to parse query: %w", err)
    }

    // Apply the parsed query to GORM
    // The rql library typically returns SQL and args
    // We need to adapt this to GORM's Where clause
    db = db.Where(query.SQL, query.Args...)

    return db, nil
}
```

#### Story 1.3: Injection Prevention

**Built-in with parameterized queries:**
```go
// The rql library and GORM both use parameterized queries by default
// Example of safe query building:

db = db.Where("email = ?", userInput)  // Safe - parameterized
// NOT: db.Where("email = " + userInput) // NEVER do this - SQL injection!
```

### Phase 2: Security & ACL Integration (The "Hemlis" Protocol)

#### Story 2.1: The "Dual-Constraint" Query

**Implementation:**
```go
// src/filter/acl.go
package filter

import (
    "gorm.io/gorm"
)

// ApplyACLConstraints adds permission checks to the query
func ApplyACLConstraints(db *gorm.DB, resourceType string, memberID *int64) *gorm.DB {
    switch resourceType {
    case "entry":
        // Entries with no permissions are public
        // Entries with user_id=0 in permissions are secret to everyone (visible)
        // Entries with specific user_ids are personal secrets (only visible to those users)

        if memberID == nil {
            // Unauthenticated: only show entries that are public OR secret-to-everyone
            db = db.Where(`
                id NOT IN (
                    SELECT id FROM cl2003_permissions
                    WHERE user_id != 0
                )
            `)
        } else {
            // Authenticated: show public, secret-to-everyone, OR entries they have permission for
            db = db.Where(`
                id NOT IN (
                    SELECT id FROM cl2003_permissions
                    WHERE user_id != 0 AND user_id != ?
                )
            `, *memberID)
        }

    case "article":
        // Articles require read:article scope (checked at middleware level)
        // No row-level ACL for articles

    // Add more resource types as needed
    }

    return db
}

// QueryWithFiltersAndACL is the main entry point combining user filters and ACL
func QueryWithFiltersAndACL(
    db *gorm.DB,
    schema *Schema,
    queryString string,
    userScopes []string,
    memberID *int64,
    resourceType string,
) (*gorm.DB, error) {

    // Step 1: Apply user filters (validated against schema)
    parser := NewRQLParser(schema)
    db, err := parser.ParseAndApply(db, queryString, userScopes)
    if err != nil {
        return nil, err
    }

    // Step 2: Apply ACL constraints (MANDATORY - cannot be bypassed)
    db = ApplyACLConstraints(db, resourceType, memberID)

    return db, nil
}
```

#### Story 2.2: Conditional Data Masking

**Implementation:**
```go
// src/filter/masking.go
package filter

// ApplyConditionalMasking handles "hemlis" masking at database level
// This uses CASE WHEN in SELECT to avoid fetching sensitive data

func ApplyEntryMasking(db *gorm.DB, memberID *int64) *gorm.DB {
    if memberID == nil {
        // Unauthenticated: mask personal secrets
        db = db.Select(`
            id, date, time, datetime, status, cl, sig, email, place,
            CASE
                WHEN id IN (
                    SELECT id FROM cl2003_permissions WHERE user_id != 0
                )
                THEN 'hemlis'
                ELSE msg
            END as msg,
            CASE
                WHEN id IN (SELECT id FROM cl2003_permissions WHERE user_id != 0)
                THEN 0
                ELSE likes
            END as likes,
            secret, personal_secret
        `)
    } else {
        // Authenticated: mask only entries they don't have permission for
        db = db.Select(`
            id, date, time, datetime, status, cl, sig, email, place,
            CASE
                WHEN id IN (
                    SELECT id FROM cl2003_permissions
                    WHERE user_id != 0 AND user_id != ?
                )
                THEN 'hemlis'
                ELSE msg
            END as msg,
            CASE
                WHEN id IN (
                    SELECT id FROM cl2003_permissions
                    WHERE user_id != 0 AND user_id != ?
                )
                THEN 0
                ELSE likes
            END as likes,
            secret, personal_secret
        `, *memberID, *memberID)
    }

    return db
}
```

### Phase 3: Prevention of Side-Channel Leaks

#### Story 3.1: Context-Aware Filtering Rules

**Implementation:**
```go
// src/filter/context_aware.go
package filter

// ContextAwareFieldValidator validates fields based on what user can actually see
type ContextAwareFieldValidator struct {
    schema   *Schema
    memberID *int64
}

func NewContextAwareValidator(schema *Schema, memberID *int64) *ContextAwareFieldValidator {
    return &ContextAwareFieldValidator{
        schema:   schema,
        memberID: memberID,
    }
}

// ValidateFilterField checks if user is allowed to filter by this field
func (v *ContextAwareFieldValidator) ValidateFilterField(publicField string, userScopes []string) error {
    field, err := v.schema.ValidateField(publicField, userScopes)
    if err != nil {
        return err
    }

    // Special case: if filtering on "msg" field for entries
    if field.DBColumn == "msg" {
        // Users cannot filter by message content if they would see "hemlis"
        // This prevents side-channel attacks like ?q=msg==password

        // Only allow if user has explicit permission or is querying their own entries
        if v.memberID == nil {
            return fmt.Errorf("filtering by message content requires authentication")
        }

        // Could add more granular checks here based on entry ownership
    }

    return nil
}

// SanitizeQueryString removes dangerous filter attempts
func (v *ContextAwareFieldValidator) SanitizeQueryString(queryString string, userScopes []string) (string, error) {
    // Parse the query to extract field names
    // For each field, validate it's allowed
    // Return error if any disallowed field is used

    // This would require parsing the RSQL/RQL string
    // The rql library can help with this

    return queryString, nil
}
```

### Phase 4: Advanced Data Retrieval (Aggregates)

#### Story 4.1: Virtual Field Mapping

**Implementation:**
```go
// src/filter/virtual_fields.go
package filter

// VirtualField defines a computed/aggregated field
type VirtualField struct {
    PublicName    string
    SQLExpression string // SQL to compute this field
    JoinRequired  string // Optional: table join needed
}

// ExtendedSchema includes virtual fields
type ExtendedSchema struct {
    Schema
    VirtualFields map[string]VirtualField
}

var ExtendedEntrySchema = ExtendedSchema{
    Schema: EntrySchema,
    VirtualFields: map[string]VirtualField{
        "likeCount": {
            PublicName: "likeCount",
            SQLExpression: `(
                SELECT COUNT(*)
                FROM 2003_likes
                WHERE 2003_likes.id = cl2003_msgs.id
            )`,
        },
        "hasPermissions": {
            PublicName: "hasPermissions",
            SQLExpression: `(
                SELECT COUNT(*) > 0
                FROM cl2003_permissions
                WHERE cl2003_permissions.id = cl2003_msgs.id
            )`,
        },
    },
}

// ApplyVirtualField adds a virtual field to the query
func (s *ExtendedSchema) ApplyVirtualField(db *gorm.DB, fieldName string) *gorm.DB {
    vf, exists := s.VirtualFields[fieldName]
    if !exists {
        return db
    }

    // Add the virtual field as a subquery in SELECT
    db = db.Select("*, " + vf.SQLExpression + " as " + fieldName)

    // If filtering by this field, use HAVING clause
    // This would be handled in the parser

    return db
}
```

## Alternative: Custom RSQL Parser (If Libraries Don't Work)

If the existing libraries don't meet your needs, here's a simple custom parser:

```go
// src/filter/custom_parser.go
package filter

import (
    "fmt"
    "strings"
    "regexp"
)

// CustomRSQLParser parses basic RSQL syntax
type CustomRSQLParser struct {
    schema *Schema
}

// Operators supported
var operators = map[string]string{
    "==": "=",
    "!=": "!=",
    "=gt=": ">",
    "=ge=": ">=",
    "=lt=": "<",
    "=le=": "<=",
    "=like=": "LIKE",
}

// ParseRSQL parses RSQL query string into SQL WHERE clause
func (p *CustomRSQLParser) ParseRSQL(query string) (string, []interface{}, error) {
    if query == "" {
        return "", nil, nil
    }

    // Split by logical operators
    // ; = AND
    // , = OR

    // Simple regex to match: field operator value
    re := regexp.MustCompile(`(\w+)(==|!=|=gt=|=ge=|=lt=|=le=|=like=)([^;,]+)`)

    matches := re.FindAllStringSubmatch(query, -1)
    if len(matches) == 0 {
        return "", nil, fmt.Errorf("invalid query syntax")
    }

    var conditions []string
    var args []interface{}

    for _, match := range matches {
        field := match[1]
        op := match[2]
        value := strings.TrimSpace(match[3])

        // Translate field
        dbColumn, err := p.schema.TranslateField(field)
        if err != nil {
            return "", nil, err
        }

        // Translate operator
        sqlOp, exists := operators[op]
        if !exists {
            return "", nil, fmt.Errorf("unknown operator: %s", op)
        }

        // Build condition with parameterized query
        condition := fmt.Sprintf("%s %s ?", dbColumn, sqlOp)
        conditions = append(conditions, condition)
        args = append(args, value)
    }

    // Combine with AND (for now)
    whereClause := strings.Join(conditions, " AND ")

    return whereClause, args, nil
}
```

## Integration with Existing Codebase

**Example: Adding to Entry Handler**

```go
// src/router/entry.go (modified)

func (eh EntryHandler) readAllEntryHandler(w http.ResponseWriter, r *http.Request) {
    take := MakeDefaultInt(r, "take", "20")
    skip := MakeDefaultInt(r, "skip", "0")

    // NEW: Get query parameter
    queryString := r.URL.Query().Get("q")

    // Get member from context (for ACL)
    member := auth.GetMember(r)
    var memberID *int64
    if member != nil {
        memberID = &member.Number
    }

    // Get user scopes
    scopes := auth.GetScopes(r)

    // Start with base query
    db := eh.db.(*gorm.DB).Model(&models.Entry{})

    // NEW: Apply filters and ACL
    db, err := filter.QueryWithFiltersAndACL(
        db,
        &filter.EntrySchema,
        queryString,
        scopes,
        memberID,
        "entry",
    )
    if err != nil {
        http.Error(w, fmt.Sprintf("Invalid filter: %v", err), http.StatusBadRequest)
        return
    }

    // Apply masking
    db = filter.ApplyEntryMasking(db, memberID)

    // Execute query
    var entries []models.Entry
    db = db.Limit(take).Offset(skip).Order("id DESC").Find(&entries)

    if db.Error != nil {
        w.WriteHeader(http.StatusInternalServerError)
        http.Error(w, fmt.Sprintf("Database error: %v", db.Error), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(entries)
}
```

## Testing Strategy

```go
// src/filter/parser_test.go
func TestRSQLParsing(t *testing.T) {
    tests := []struct{
        name string
        query string
        expectedSQL string
        expectedArgs []interface{}
        shouldError bool
    }{
        {
            name: "simple equality",
            query: "email==test@example.com",
            expectedSQL: "email = ?",
            expectedArgs: []interface{}{"test@example.com"},
        },
        {
            name: "greater than",
            query: "likes=gt=10",
            expectedSQL: "likes > ?",
            expectedArgs: []interface{}{10},
        },
        {
            name: "SQL injection attempt",
            query: "email=='; DROP TABLE users; --",
            // Should be parameterized, not executed
            expectedSQL: "email = ?",
            expectedArgs: []interface{}{"'; DROP TABLE users; --"},
        },
    }

    // Run tests
}
```

## Recommended Implementation Order

1. **Week 1-2: Foundation**
   - Implement Schema mapping (Story 1.1)
   - Integrate libatomic/rql library (Story 1.2)
   - Add comprehensive unit tests

2. **Week 3-4: Security**
   - Implement ACL integration (Story 2.1)
   - Add conditional masking (Story 2.2)
   - Security audit and penetration testing

3. **Week 5: Side-Channel Prevention**
   - Context-aware filtering (Story 3.1)
   - Add integration tests for attack scenarios

4. **Week 6: Advanced Features**
   - Virtual fields (Story 4.1)
   - Performance optimization
   - Documentation

## Key Security Principles

1. **Whitelist, Never Blacklist**: Only allow explicitly defined fields
2. **Dual Constraints**: User filter AND ACL check (never OR)
3. **Parameterized Queries**: Always use `?` placeholders
4. **Fail Secure**: Unknown fields → 400 Bad Request, not silent ignore
5. **Database-Level Masking**: Use CASE WHEN in SELECT, don't fetch then mask
6. **Context-Aware**: What you can filter depends on what you can see

## Dependencies to Add

```bash
go get github.com/libatomic/rql
```

