package adapter

import (
	"testing"
)

func TestRoundCPUMillicores(t *testing.T) {
	adapter := &eksAdapter{}

	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		// The problematic case that triggered this fix
		{"1024m rounds to 1000m", 1024, 1000},

		// Edge cases around quarters
		{"1000m stays 1000m", 1000, 1000},
		{"1125m rounds to 1250m", 1125, 1250},
		{"1150m rounds to 1250m", 1150, 1250},
		{"1250m stays 1250m", 1250, 1250},

		// Test rounding up and down
		{"100m rounds to 0m", 100, 0},
		{"125m rounds to 250m", 125, 250},
		{"137m rounds to 250m", 137, 250},
		{"250m stays 250m", 250, 250},
		{"374m rounds to 250m", 374, 250},
		{"375m rounds to 500m", 375, 500},
		{"500m stays 500m", 500, 500},
		{"624m rounds to 500m", 624, 500},
		{"625m rounds to 750m", 625, 750},
		{"750m stays 750m", 750, 750},

		// Higher values - test both rounding up and down
		{"2048m rounds to 2000m", 2048, 2000},
		{"2100m rounds to 2000m", 2100, 2000},
		{"2126m rounds UP to 2250m", 2126, 2250},
		{"3000m stays 3000m", 3000, 3000},
		{"3001m rounds to 3000m", 3001, 3000},
		{"3126m rounds UP to 3250m", 3126, 3250},
		{"3200m rounds UP to 3250m", 3200, 3250},

		// Large values
		{"60000m stays 60000m", 60000, 60000},
		{"60024m rounds to 60000m", 60024, 60000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.roundCPUMillicores(tt.input)
			if result != tt.expected {
				t.Errorf("roundCPUMillicores(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRoundCPUAvoidsCgroupIssue verifies that rounded values avoid the systemd
// cgroup rounding issue where non-integer percentages get rounded up by systemd
func TestRoundCPUAvoidsCgroupIssue(t *testing.T) {
	adapter := &eksAdapter{}

	// Test values that would cause systemd rounding issues
	problematicValues := []int64{
		1024, // 102.4% -> systemd rounds to 103%
		1025, // 102.5% -> systemd rounds to 103%
		1026, // 102.6% -> systemd rounds to 103%
		2048, // 204.8% -> systemd rounds to 205%
		3072, // 307.2% -> systemd rounds to 308%
	}

	for _, input := range problematicValues {
		result := adapter.roundCPUMillicores(input)

		// Verify result is a multiple of 250 (quarter core)
		if result%250 != 0 {
			t.Errorf("roundCPUMillicores(%d) = %d, which is not a multiple of 250m", input, result)
		}

		// Verify result produces an integer percentage (whole or quarter)
		// Valid: 0%, 25%, 50%, 75%, 100%, 125%, etc.
		// 1000m = 100%, 250m = 25%
		percentage := (result * 100) / 1000 // percentage with 1 decimal place
		if percentage%25 != 0 {
			t.Errorf("roundCPUMillicores(%d) = %d, which produces non-quarter percentage (%d)",
				input, result, percentage)
		}
	}
}
