# RSQL Filtering Implementation

## Overview

This implementation adds advanced RSQL (RESTful Service Query Language) filtering capabilities to the `/db/entries` endpoint with integrated access control and security features.

## Features

### 1. RSQL Query Syntax
The API supports standard RSQL operators:

**Comparison Operators:**
- `==` - Equal to
- `!=` - Not equal to
- `=gt=` - Greater than
- `=ge=` - Greater than or equal
- `=lt=` - Less than
- `=le=` - Less than or equal
- `=like=` - Pattern matching (SQL LIKE)
- `=in=` - In list
- `=out=` - Not in list

**Logical Operators:**
- `;` - AND
- `,` - OR
- `()` - Grouping

### 2. Security Controls

#### Authentication Requirement
- The `q` parameter **requires JWT authentication**
- Unauthenticated requests with `q` parameter return `401 Unauthorized`
- This prevents unauthorized users from probing the database

#### Field-Level Access Control
Fields are categorized by access level:

**Public Fields** (no authentication required):
- `id`, `date`, `time`, `datetime`
- `status`, `cl`, `place`
- `olsug`, `enheter`, `report`
- Virtual fields: `likes`, `secret`, `personal_secret`

**Sensitive Fields** (require authentication AND permission):
- `msg` - Message content
- `sig` - Signature
- `email` - Email address

Attempting to filter on sensitive fields without proper authentication returns `400 Bad Request`.

#### Permission-Based Content Masking
The existing "hemlis" system continues to work:
- Entries with permissions are filtered in results
- Unauthorized viewers see "hemlis" instead of actual content
- Message filtering applies AFTER query filtering

### 3. Virtual Fields

Virtual/computed fields can be used in filters:

**`likes`** - Number of likes (aggregated from `2003_likes` table):
```
?q=likes=gt=10        # Entries with more than 10 likes
?q=likes=le=5         # Entries with 5 or fewer likes
```

**`secret`** - Boolean flag indicating if entry has ANY permissions:
```
?q=secret==true       # Only secret entries
?q=secret==false      # Only public entries
```

**`personal_secret`** - Boolean flag for personal secrets (user_id != 0):
```
?q=personal_secret==true   # Only personal secrets
```

## API Usage

### Basic Examples

**Filter by single field:**
```bash
GET /db/entries?q=status==1
```

**Filter with AND logic:**
```bash
GET /db/entries?q=status==1;cl==5
```

**Filter with OR logic:**
```bash
GET /db/entries?q=cl==5,cl==7
```

**Pattern matching:**
```bash
GET /db/entries?q=place=like=Stockholm
```

**Numeric comparisons:**
```bash
GET /db/entries?q=id=gt=1000;id=lt=2000
```

### Complex Examples

**Grouped conditions:**
```bash
GET /db/entries?q=(status==1;cl==5),place=like=Stockholm
# Equivalent to: (status==1 AND cl==5) OR place LIKE '%Stockholm%'
```

**List membership:**
```bash
GET /db/entries?q=cl=in=(5,7,9)
# Equivalent to: cl IN (5, 7, 9)
```

**Multiple criteria:**
```bash
GET /db/entries?q=status==1;likes=gt=10;secret==false
# Public entries with status 1 and more than 10 likes
```

### Pagination with Filtering

Combine `q` with pagination parameters:
```bash
GET /db/entries?q=status==1&take=50&skip=100
```

## Implementation Details

### Architecture

```
┌─────────────────┐
│  HTTP Request   │
│  ?q=status==1   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Entry Handler   │ ◄─── Authentication Check (JWT required)
│  (router)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ RSQL Adapter    │ ◄─── Field whitelist check
│  (gorm_adapter) │      Permission validation
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ go-rsql Parser  │ ◄─── Syntax parsing
│  (library)      │      Operator translation
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ SQL Generation  │ ◄─── Virtual field subqueries
│                 │      SQL injection prevention
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ GORM/MySQL      │
│  WHERE clause   │
└─────────────────┘
```

### SQL Injection Prevention

All user input is safely escaped and parameterized:
- Single quotes escaped: `'` → `''`
- Values passed as query parameters, not concatenated
- Field names validated against whitelist before use

Example transformation:
```
Input:  q=msg=like=Robert'); DROP TABLE--
Output: msg LIKE '%Robert''); DROP TABLE--%'
        (harmless string search, not SQL injection)
```

### Virtual Field Implementation

Virtual fields use SQL subqueries:

**Likes count:**
```sql
WHERE (SELECT COUNT(*) FROM 2003_likes 
       WHERE 2003_likes.id = cl2003_msgs.id) > 10
```

**Secret flag:**
```sql
WHERE EXISTS (SELECT 1 FROM cl2003_permissions 
              WHERE cl2003_permissions.id = cl2003_msgs.id)
```

## Error Responses

**401 Unauthorized** - Unauthenticated user trying to use `q` parameter:
```json
{"error": "authentication required to use query filters (q parameter)"}
```

**400 Bad Request** - Invalid field or insufficient permissions:
```json
{"error": "field 'msg' cannot be used in filter due to insufficient permissions"}
```

**400 Bad Request** - Invalid RSQL syntax:
```json
{"error": "invalid filter: unexpected token at position 5"}
```

## Testing

Run the test suite:
```bash
go test ./src/rsql/...
```

All tests validate:
- Field mapping and whitelisting
- Permission-based field access
- RSQL parsing for all operators
- Error handling for invalid queries

## Performance Considerations

1. **Indexes**: Ensure database indexes exist on frequently filtered columns
2. **Virtual Fields**: Subqueries can be expensive on large datasets
3. **Pagination**: Always use `take` and `skip` to limit result sets
4. **Complex Queries**: Deeply nested conditions may impact performance

## Future Enhancements

Potential improvements:
- [ ] Query result caching
- [ ] Custom field aliases
- [ ] Full-text search integration
- [ ] Query complexity limits
- [ ] Query performance metrics

## References

- [RSQL Specification](https://github.com/jirutka/rsql-parser)
- [go-rsql Library](https://github.com/rbicker/go-rsql)
- [FIQL (Feed Item Query Language)](https://tools.ietf.org/html/draft-nottingham-atompub-fiql-00)
