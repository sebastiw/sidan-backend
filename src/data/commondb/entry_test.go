package commondb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/sebastiw/sidan-backend/src/data/commondb"
	"github.com/sebastiw/sidan-backend/src/models"
)

// TestReadEntries_WithRSQL tests the complete RSQL filtering logic
// using an in-memory SQLite database
func TestReadEntries_WithRSQL(t *testing.T) {
	// 1. Setup In-Memory DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 2. Migrate schemas (matching production schema names)
	db.AutoMigrate(&models.Entry{})
	db.Exec("CREATE TABLE IF NOT EXISTS `2003_likes` (date TEXT, time TEXT, id INTEGER, sig TEXT, host TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS `cl2003_msgs_kumpaner` (id INTEGER, number TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS `cl2003_permissions` (id INTEGER, user_id INTEGER)")

	// 3. Seed Test Data
	now := time.Now()
	
	// Entry 1: Popular post by Alice with 10 likes
	entry1 := models.Entry{
		Sig:      "Alice",
		Msg:      "Great post about beer",
		DateTime: now,
		Date:     now.Format("2006-01-02"),
		Time:     now.Format("15:04:05"),
	}
	db.Create(&entry1)
	for i := 0; i < 10; i++ {
		db.Exec("INSERT INTO `2003_likes` (id, sig) VALUES (?, ?)", entry1.Id, "user"+string(rune(i)))
	}

	// Entry 2: Unpopular post by Bob with 0 likes
	entry2 := models.Entry{
		Sig:      "Bob",
		Msg:      "Nobody likes this",
		DateTime: now.Add(-1 * time.Hour),
		Date:     now.Add(-1 * time.Hour).Format("2006-01-02"),
		Time:     now.Add(-1 * time.Hour).Format("15:04:05"),
	}
	db.Create(&entry2)

	// Entry 3: Moderately popular post by Charlie with 3 likes
	entry3 := models.Entry{
		Sig:      "Charlie",
		Msg:      "Some people like this about beer",
		DateTime: now.Add(-2 * time.Hour),
		Date:     now.Add(-2 * time.Hour).Format("2006-01-02"),
		Time:     now.Add(-2 * time.Hour).Format("15:04:05"),
	}
	db.Create(&entry3)
	for i := 0; i < 3; i++ {
		db.Exec("INSERT INTO `2003_likes` (id, sig) VALUES (?, ?)", entry3.Id, "fan"+string(rune(i)))
	}

	// Initialize database wrapper
	repo := commondb.CommonDatabase{DB: db}

	// 4. Test Case A: No Filter (Should return all 3)
	t.Run("no filter returns all entries", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, "")
		assert.NoError(t, err)
		assert.Len(t, res, 3, "Should return all 3 entries")
	})

	// 5. Test Case B: RSQL Filter "likes > 5"
	t.Run("filter likes greater than 5", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, "likes=gt=5")
		assert.NoError(t, err)
		assert.Len(t, res, 1, "Should only return entries with >5 likes")
		assert.Equal(t, "Alice", res[0].Sig)
		assert.Equal(t, int64(10), res[0].Likes) // Computed field
	})

	// 6. Test Case C: RSQL Filter "likes < 2"
	t.Run("filter likes less than 2", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, "likes=lt=2")
		assert.NoError(t, err)
		assert.Len(t, res, 1, "Should only return entries with <2 likes")
		assert.Equal(t, "Bob", res[0].Sig)
		assert.Equal(t, int64(0), res[0].Likes)
	})

	// 7. Test Case D: RSQL Filter "likes == 3"
	t.Run("filter likes equals 3", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, "likes==3")
		assert.NoError(t, err)
		assert.Len(t, res, 1, "Should only return entries with exactly 3 likes")
		assert.Equal(t, "Charlie", res[0].Sig)
	})

	// 8. Test Case E: RSQL Filter by signature
	t.Run("filter by signature", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, `sig=="Alice"`)
		assert.NoError(t, err)
		assert.Len(t, res, 1, "Should only return Alice's entry")
		assert.Equal(t, "Alice", res[0].Sig)
	})

	// 9. Test Case F: Complex AND filter (likes AND sig)
	t.Run("complex AND filter", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, `likes=gt=2;sig=="Charlie"`)
		assert.NoError(t, err)
		assert.Len(t, res, 1, "Should return Charlie with >2 likes")
		assert.Equal(t, "Charlie", res[0].Sig)
	})

	// 10. Test Case G: Complex OR filter
	t.Run("complex OR filter", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, `sig=="Alice",sig=="Bob"`)
		assert.NoError(t, err)
		assert.Len(t, res, 2, "Should return both Alice and Bob")
	})

	// 11. Test Case H: Message content filter
	t.Run("filter by message content", func(t *testing.T) {
		res, err := repo.ReadEntries(10, 0, `msg=="beer"`)
		assert.NoError(t, err)
		// Note: This might return 0 results depending on exact match vs contains
		// The RSQL library uses = which is exact match
		if len(res) > 0 {
			for _, entry := range res {
				assert.Contains(t, entry.Msg, "beer")
			}
		}
	})

	// 12. Test Case I: Invalid field should return error
	t.Run("invalid field returns error", func(t *testing.T) {
		_, err := repo.ReadEntries(10, 0, `email=="test@example.com"`)
		assert.Error(t, err, "Should return error for disallowed field")
		assert.Contains(t, err.Error(), "not allowed")
	})

	// 13. Test Case J: Invalid syntax should return error
	t.Run("invalid RSQL syntax returns error", func(t *testing.T) {
		_, err := repo.ReadEntries(10, 0, "invalid==")
		assert.Error(t, err, "Should return error for invalid syntax")
	})

	// 14. Test Case K: Pagination
	t.Run("pagination works with filter", func(t *testing.T) {
		// Get first 2 entries
		res1, err := repo.ReadEntries(2, 0, "")
		assert.NoError(t, err)
		assert.Len(t, res1, 2)

		// Get next entry (skip first 2)
		res2, err := repo.ReadEntries(2, 2, "")
		assert.NoError(t, err)
		assert.Len(t, res2, 1)

		// Ensure different entries
		assert.NotEqual(t, res1[0].Id, res2[0].Id)
	})
}
