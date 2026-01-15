package rsql

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

// CanViewerSeeSecretContent determines if a viewer can see secret entry content
// This is used to decide if they can filter on sensitive fields like msg, sig, email
// Returns true if:
// - Entry has no permissions (public)
// - Entry is secret to everyone (user_id=0)
// - Viewer is authenticated AND (is author OR has explicit permission)
func CanViewerSeeSecretContent(entry *models.Entry, viewerMemberID *int64) bool {
	// No permissions = public entry
	if len(entry.Permissions) == 0 {
		return true
	}
	
	// Check if secret to everyone (user_id=0)
	isSecretToEveryone := false
	permittedUserIDs := []int64{}
	
	for _, perm := range entry.Permissions {
		if perm.UserId == 0 {
			isSecretToEveryone = true
		} else {
			permittedUserIDs = append(permittedUserIDs, perm.UserId)
		}
	}
	
	// Secret to everyone → anyone can see
	if isSecretToEveryone {
		return true
	}
	
	// Personal secret - check if viewer has permission
	if viewerMemberID == nil {
		return false // Unauthenticated users cannot see personal secrets
	}
	
	// Check if viewer is in permitted list
	for _, permUserID := range permittedUserIDs {
		if permUserID == *viewerMemberID {
			return true
		}
	}
	
	// Check if viewer is the author
	// Note: This requires parsing the sig field, which we handle in entry_filter.go
	// For filtering purposes, we'll be conservative and require explicit permission
	
	return false
}

// CanViewerFilterOnSensitiveFields determines if viewer can use sensitive fields in filters
// This prevents side-channel attacks where users could guess hidden content by filtering
// Returns true if viewer is authenticated (we assume they might have access to SOME entries)
func CanViewerFilterOnSensitiveFields(viewerMemberID *int64) bool {
	// Conservative approach: only authenticated users can filter on sensitive fields
	// This prevents unauthenticated users from using filters to probe for content
	return viewerMemberID != nil
}
