package rsql

import (
	"testing"
)

func TestGetFieldMapping(t *testing.T) {
	tests := []struct {
		field    string
		expected bool
	}{
		{"id", true},
		{"msg", true},
		{"likes", true},
		{"invalid_field", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			mapping := GetFieldMapping(tt.field)
			exists := mapping != nil
			if exists != tt.expected {
				t.Errorf("Field '%s': expected exists=%v, got exists=%v", tt.field, tt.expected, exists)
			}
		})
	}
}

func TestIsFieldAllowed(t *testing.T) {
	memberID := int64(123)
	
	tests := []struct {
		name                 string
		field                string
		viewerMemberID       *int64
		canViewSecretContent bool
		expected             bool
	}{
		{"Public field, no auth", "id", nil, false, true},
		{"Public field, with auth", "id", &memberID, false, true},
		{"Sensitive field, no auth, can't view", "msg", nil, false, false},
		{"Sensitive field, no auth, can view", "msg", nil, true, false}, // Still requires auth
		{"Sensitive field, with auth, can't view", "msg", &memberID, false, false},
		{"Sensitive field, with auth, can view", "msg", &memberID, true, true},
		{"Invalid field", "invalid", &memberID, true, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := IsFieldAllowed(tt.field, tt.viewerMemberID, tt.canViewSecretContent)
			if allowed != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, allowed)
			}
		})
	}
}

func TestGormAdapter_BasicParsing(t *testing.T) {
	memberID := int64(123)
	ctx := FilterContext{
		ViewerMemberID:       &memberID,
		CanViewSecretContent: true,
	}
	
	adapter, err := NewGormAdapter(ctx)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	
	tests := []struct {
		name     string
		query    string
		wantErr  bool
	}{
		{"Simple equality", "status==1", false},
		{"Greater than", "id=gt=100", false},
		{"Less than", "id=lt=200", false},
		{"Like", "place=like=Stockholm", false},
		{"AND condition", "status==1;cl==5", false},
		{"OR condition", "cl==5,cl==7", false},
		{"Complex with parens", "(status==1;cl==5),place=like=Stockholm", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't run actual database queries in unit tests,
			// but we can test that parsing doesn't error
			_, err := adapter.parser.Process(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGormAdapter_FieldPermissions(t *testing.T) {
	// Unauthenticated user
	ctx := FilterContext{
		ViewerMemberID:       nil,
		CanViewSecretContent: false,
	}
	
	adapter, err := NewGormAdapter(ctx)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	
	// Try to filter on sensitive field (should fail)
	result, err := adapter.parser.Process("msg==secret")
	if err != nil {
		t.Fatalf("Parser failed: %v", err)
	}
	
	// Result should contain error marker
	if result != `{"error":"field 'msg' not allowed"}` {
		t.Errorf("Expected error marker in result, got: %s", result)
	}
}

func TestCanViewerFilterOnSensitiveFields(t *testing.T) {
	memberID := int64(123)
	
	tests := []struct {
		name           string
		viewerMemberID *int64
		expected       bool
	}{
		{"Authenticated user", &memberID, true},
		{"Unauthenticated user", nil, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanViewerFilterOnSensitiveFields(tt.viewerMemberID)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
