package state

import "testing"

func TestPoolTier(t *testing.T) {
	tests := []struct {
		name     string
		cpu      int64
		mem      int64
		expected string
	}{
		{"default small", 256, 512, "s"},
		{"max small cpu", 2999, 23999, "s"},
		{"cpu triggers medium", 3000, 512, "m"},
		{"mem triggers medium", 256, 24000, "m"},
		{"max medium cpu", 13999, 99999, "m"},
		{"cpu triggers large", 14000, 512, "l"},
		{"mem triggers large", 256, 100000, "l"},
		{"max large cpu", 29999, 199999, "l"},
		{"cpu triggers xl", 30000, 512, "xl"},
		{"mem triggers xl", 256, 200000, "xl"},
		{"both large", 60000, 350000, "xl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PoolTier(tt.cpu, tt.mem)
			if got != tt.expected {
				t.Errorf("PoolTier(%d, %d) = %q, want %q", tt.cpu, tt.mem, got, tt.expected)
			}
		})
	}
}
