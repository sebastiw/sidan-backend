package router

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sebastiw/sidan-backend/src/models"
)

// FilterEntryMessage applies permission-based message filtering
// Logic matches stored procedure ReadEntries:
// - No permissions (public) → show full message
// - user_id=0 (secret to everyone) → show full message
// - Has specific user_ids and (requester in list OR requester is author) → show message with prefix
// - Has specific user_ids and requester NOT in list → show only "hemlis" and clear all other fields
func FilterEntryMessage(entry *models.Entry, viewerMemberID *int64) {
	// No permissions = public entry, show full message
	if len(entry.Permissions) == 0 {
		return
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

	// Secret to everyone (user_id=0) → show full message
	if isSecretToEveryone {
		return
	}

	// Personal secret - check if viewer has permission
	// No viewer (unauthenticated) → show "hemlis" and clear all fields
	if viewerMemberID == nil {
		redactEntry(entry)
		return
	}

	// Check if viewer is the author (extract member number from sig like "#8")
	authorMemberID := extractMemberIDFromSig(entry.Sig)
	isAuthor := authorMemberID != nil && *authorMemberID == *viewerMemberID

	// Check if viewer is in permitted list
	isPermitted := false
	for _, permUserID := range permittedUserIDs {
		if permUserID == *viewerMemberID {
			isPermitted = true
			break
		}
	}

	// Viewer is author OR in permitted list → show message with prefix
	if isAuthor || isPermitted {
		// Build "hemlis Till #1,#2,#3" prefix
		userIDStrings := make([]string, len(permittedUserIDs))
		for i, uid := range permittedUserIDs {
			userIDStrings[i] = fmt.Sprintf("#%d", uid)
		}
		prefix := fmt.Sprintf("<small>hemlis Till %s:</small><br>", strings.Join(userIDStrings, ","))
		entry.Msg = prefix + entry.Msg
		return
	}

	// Viewer NOT authorized → show only "hemlis" and clear all fields
	redactEntry(entry)
}

// redactEntry clears all sensitive fields from an entry, leaving only "hemlis"
// Clears: msg (→ "hemlis"), sig, email, place, ip, host, lat, lon, sidekicks, likes
func redactEntry(entry *models.Entry) {
	entry.Msg = "hemlis"
	entry.Sig = ""
	entry.Email = ""
	entry.Place = ""
	entry.Ip = nil
	entry.Host = nil
	entry.Lat = nil
	entry.Lon = nil
	entry.Olsug = 0
	entry.Enheter = 0
	entry.SideKicks = nil
	entry.Likes = 0
}

// FilterEntriesMessages applies FilterEntryMessage to a slice of entries
func FilterEntriesMessages(entries []models.Entry, viewerMemberID *int64) {
	for i := range entries {
		FilterEntryMessage(&entries[i], viewerMemberID)
	}
}

// extractMemberIDFromSig extracts member ID from signature like "#123" or "P456"
// Returns nil if signature doesn't match expected format
func extractMemberIDFromSig(sig string) *int64 {
	if len(sig) < 2 {
		return nil
	}

	// Handle "#123", "P456", "S789" formats
	if sig[0] == '#' || sig[0] == 'P' || sig[0] == 'S' {
		numStr := sig[1:]
		if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return &num
		}
	}

	return nil
}
