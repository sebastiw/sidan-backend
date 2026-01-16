package filter

import (
	"fmt"

	"gorm.io/gorm"
)

// ApplyConditionalMasking handles "hemlis" masking at database level
// This uses CASE WHEN in SELECT to avoid fetching sensitive data into memory
// This is critical for security - we never fetch sensitive data if user shouldn't see it
func ApplyConditionalMasking(db *gorm.DB, resourceType string, memberID *int64) *gorm.DB {
	if resourceType == "entry" {
		return applyEntryMasking(db, memberID)
	}
	return db
}

// applyEntryMasking applies database-level masking for entry messages
func applyEntryMasking(db *gorm.DB, memberID *int64) *gorm.DB {
	if memberID == nil {
		// Unauthenticated: mask personal secrets
		db = db.Select(`
			cl2003_msgs.id,
			cl2003_msgs.date,
			cl2003_msgs.time,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions
					WHERE cl2003_permissions.id = cl2003_msgs.id
					AND cl2003_permissions.user_id != 0
				)
				THEN 'hemlis'
				ELSE cl2003_msgs.msg
			END as msg,
			cl2003_msgs.status,
			cl2003_msgs.cl,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions
					WHERE cl2003_permissions.id = cl2003_msgs.id
					AND cl2003_permissions.user_id != 0
				)
				THEN ''
				ELSE cl2003_msgs.sig
			END as sig,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions
					WHERE cl2003_permissions.id = cl2003_msgs.id
					AND cl2003_permissions.user_id != 0
				)
				THEN ''
				ELSE cl2003_msgs.email
			END as email,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions
					WHERE cl2003_permissions.id = cl2003_msgs.id
					AND cl2003_permissions.user_id != 0
				)
				THEN ''
				ELSE cl2003_msgs.place
			END as place,
			cl2003_msgs.ip,
			cl2003_msgs.host,
			cl2003_msgs.olsug,
			cl2003_msgs.enheter,
			cl2003_msgs.lat,
			cl2003_msgs.lon,
			cl2003_msgs.report
		`)
	} else {
		// Authenticated: mask only entries they don't have permission for
		// Need to check if they have permission OR if they're the author
		db = db.Select(fmt.Sprintf(`
			cl2003_msgs.id,
			cl2003_msgs.date,
			cl2003_msgs.time,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions p
					WHERE p.id = cl2003_msgs.id
					AND p.user_id != 0
					AND p.user_id != %d
					AND cl2003_msgs.sig NOT LIKE '#%d%%'
				)
				THEN 'hemlis'
				ELSE cl2003_msgs.msg
			END as msg,
			cl2003_msgs.status,
			cl2003_msgs.cl,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions p
					WHERE p.id = cl2003_msgs.id
					AND p.user_id != 0
					AND p.user_id != %d
					AND cl2003_msgs.sig NOT LIKE '#%d%%'
				)
				THEN ''
				ELSE cl2003_msgs.sig
			END as sig,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions p
					WHERE p.id = cl2003_msgs.id
					AND p.user_id != 0
					AND p.user_id != %d
					AND cl2003_msgs.sig NOT LIKE '#%d%%'
				)
				THEN ''
				ELSE cl2003_msgs.email
			END as email,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM cl2003_permissions p
					WHERE p.id = cl2003_msgs.id
					AND p.user_id != 0
					AND p.user_id != %d
					AND cl2003_msgs.sig NOT LIKE '#%d%%'
				)
				THEN ''
				ELSE cl2003_msgs.place
			END as place,
			cl2003_msgs.ip,
			cl2003_msgs.host,
			cl2003_msgs.olsug,
			cl2003_msgs.enheter,
			cl2003_msgs.lat,
			cl2003_msgs.lon,
			cl2003_msgs.report
		`, *memberID, *memberID, *memberID, *memberID, *memberID, *memberID, *memberID, *memberID))
	}

	return db
}
