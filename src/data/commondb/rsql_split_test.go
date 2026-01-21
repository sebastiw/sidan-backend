package commondb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitWhereHaving(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectWhere   string
		expectHaving  string
	}{
		{
			name:         "aggregate only",
			input:        "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5",
			expectWhere:  "",
			expectHaving: "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5",
		},
		{
			name:         "regular field only",
			input:        `cl2003_msgs.sig = "Alice"`,
			expectWhere:  `cl2003_msgs.sig = "Alice"`,
			expectHaving: "",
		},
		{
			name:         "joined table field",
			input:        "SideKicks.number = 123",
			expectWhere:  "SideKicks.number = 123",
			expectHaving: "",
		},
		{
			name:         "mixed aggregate and regular",
			input:        `(COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 3 AND cl2003_msgs.sig = "Alice")`,
			expectWhere:  `cl2003_msgs.sig = "Alice"`,
			expectHaving: "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 3",
		},
		{
			name:         "multiple regular fields",
			input:        `(cl2003_msgs.sig = "Alice" AND cl2003_msgs.msg = "beer")`,
			expectWhere:  `cl2003_msgs.sig = "Alice" AND cl2003_msgs.msg = "beer"`,
			expectHaving: "",
		},
		{
			name:         "mixed with joined table",
			input:        `(SideKicks.number = 123 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5)`,
			expectWhere:  "SideKicks.number = 123",
			expectHaving: "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5",
		},
		{
			name:         "complex with multiple conditions",
			input:        `(cl2003_msgs.sig = "Alice" AND SideKicks.number = 123 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5)`,
			expectWhere:  "cl2003_msgs.sig = \"Alice\" AND SideKicks.number = 123",
			expectHaving: "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5",
		},
		{
			name:         "OR with aggregate - entire condition goes to HAVING",
			input:        `(SideKicks.number IN (7,8,9) OR (COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 15))`,
			expectWhere:  "",
			expectHaving: "SideKicks.number IN (7,8,9) OR (COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 15)",
		},
		{
			name:         "complex OR with mixed AND - all to HAVING when aggregate present",
			input:        `((SideKicks.number IN (7,8,9) AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5) OR (SideKicks.number NOT IN (7,8,9) AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 5))`,
			expectWhere:  "",
			expectHaving: "(SideKicks.number IN (7,8,9) AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5) OR (SideKicks.number NOT IN (7,8,9) AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 5)",
		},
		{
			name:         "nested OR inside AND with regular fields - should split",
			input:        `(cl2003_msgs.sig = "Alice" AND (cl2003_msgs.msg = "beer" OR cl2003_msgs.msg = "wine"))`,
			expectWhere:  `cl2003_msgs.sig = "Alice" AND (cl2003_msgs.msg = "beer" OR cl2003_msgs.msg = "wine")`,
			expectHaving: "",
		},
		{
			name:         "multiple aggregates AND condition - all to HAVING",
			input:        `(COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 10)`,
			expectWhere:  "",
			expectHaving: "COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) > 5 AND COUNT(DISTINCT CONCAT(LikeRecords.sig, '|', LikeRecords.host)) < 10",
		},
		{
			name:         "complex nested regular fields only - all to WHERE",
			input:        `((cl2003_msgs.sig = "Alice" OR cl2003_msgs.sig = "Bob") AND cl2003_msgs.msg = "test")`,
			expectWhere:  `(cl2003_msgs.sig = "Alice" OR cl2003_msgs.sig = "Bob") AND cl2003_msgs.msg = "test"`,
			expectHaving: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, having := SplitWhereHaving(tt.input)
			assert.Equal(t, tt.expectWhere, where, "WHERE clause mismatch")
			assert.Equal(t, tt.expectHaving, having, "HAVING clause mismatch")
		})
	}
}

func TestSplitOnAND(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single condition",
			input:    `sig = "Alice"`,
			expected: []string{`sig = "Alice"`},
		},
		{
			name:     "two conditions",
			input:    `sig = "Alice" AND likes > 5`,
			expected: []string{`sig = "Alice"`, `likes > 5`},
		},
		{
			name:     "three conditions",
			input:    `sig = "Alice" AND likes > 5 AND msg = "beer"`,
			expected: []string{`sig = "Alice"`, `likes > 5`, `msg = "beer"`},
		},
		{
			name:     "nested parentheses",
			input:    `(sig = "Alice" AND likes > 5) AND msg = "beer"`,
			expected: []string{`(sig = "Alice" AND likes > 5)`, `msg = "beer"`},
		},
		{
			name:     "function with AND inside",
			input:    `CONCAT(sig, ' AND ', host) = "test" AND likes > 5`,
			expected: []string{`CONCAT(sig, ' AND ', host) = "test"`, `likes > 5`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitOnAND(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
