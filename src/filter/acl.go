package filter

import (
	"gorm.io/gorm"
)

// ApplyACLConstraints adds permission checks to the query
// This ensures that user filters are ALWAYS combined with ACL checks
// Implements the "Dual-Constraint" pattern: (User Filter) AND (ACL Check)
func ApplyACLConstraints(db *gorm.DB, resourceType string, memberID *int64) *gorm.DB {
	if resourceType == "entry" {
		return applyEntryACL(db, memberID)
	}
	return db
}

// applyEntryACL applies ACL constraints for entries based on permissions table
func applyEntryACL(db *gorm.DB, memberID *int64) *gorm.DB {
	// Entry ACL logic:
	// 1. Entries with NO permissions → Public (visible to all)
	// 2. Entries with user_id=0 → Secret to everyone (visible to all, but flagged)
	// 3. Entries with specific user_ids → Personal secret (only visible to those users + author)

	if memberID == nil {
		// Unauthenticated users: Show entries that are public OR secret-to-everyone
		// Hide entries with specific user permissions (personal secrets)
		db = db.Where(`
			id NOT IN (
				SELECT id FROM cl2003_permissions
				WHERE user_id != 0
			)
		`)
	} else {
		// Authenticated users: Show entries they have permission for
		// This includes: public entries, secret-to-everyone, and personal secrets they're allowed to see
		db = db.Where(`
			id NOT IN (
				SELECT p.id
				FROM cl2003_permissions p
				WHERE p.user_id != 0
				AND p.user_id != ?
				AND p.id NOT IN (
					SELECT id FROM cl2003_msgs WHERE sig = ?
				)
			)
		`, *memberID, sigForMember(*memberID))
	}

	return db
}

// sigForMember returns the signature format for a member number (e.g., "#8")
func sigForMember(memberID int64) string {
	return "#" + string(rune(memberID+'0'))
}

// QueryWithFiltersAndACL is the main entry point combining user filters and ACL
// This is the ONLY way to query filtered data - ensures security cannot be bypassed
func QueryWithFiltersAndACL(
	db *gorm.DB,
	schema *Schema,
	queryString string,
	sortString string,
	userScopes []string,
	memberID *int64,
	resourceType string,
) (*gorm.DB, error) {

	// Step 1: Apply user filters (validated against schema)
	parser := NewFilterParser(schema, userScopes)

	var err error
	db, err = parser.ParseAndApply(db, queryString)
	if err != nil {
		return nil, err
	}

	// Step 2: Apply sorting (if provided)
	db, err = parser.ParseSorting(db, sortString)
	if err != nil {
		return nil, err
	}

	// Step 3: Apply ACL constraints (MANDATORY - cannot be bypassed)
	// This ensures that even if filters are malicious, ACL is enforced
	db = ApplyACLConstraints(db, resourceType, memberID)

	return db, nil
}
