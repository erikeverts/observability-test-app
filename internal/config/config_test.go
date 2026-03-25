package config

import (
	"testing"
	"time"
)

func TestParseErrorRoutes(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]float64
	}{
		{"", map[string]float64{}},
		{"/path:0.1", map[string]float64{"/path": 0.1}},
		{"/a:0.1,/b:0.25", map[string]float64{"/a": 0.1, "/b": 0.25}},
		{"invalid", map[string]float64{}},
		{"/path:notanumber", map[string]float64{}},
	}

	for _, tt := range tests {
		got := parseErrorRoutes(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("parseErrorRoutes(%q): got %d entries, want %d", tt.input, len(got), len(tt.expected))
			continue
		}
		for k, v := range tt.expected {
			if got[k] != v {
				t.Errorf("parseErrorRoutes(%q)[%s]: got %f, want %f", tt.input, k, got[k], v)
			}
		}
	}
}

func TestParseLatencyRoutes(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]time.Duration
	}{
		{"", map[string]time.Duration{}},
		{"/path:200ms", map[string]time.Duration{"/path": 200 * time.Millisecond}},
		{"/a:1s,/b:500ms", map[string]time.Duration{"/a": 1 * time.Second, "/b": 500 * time.Millisecond}},
	}

	for _, tt := range tests {
		got := parseLatencyRoutes(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("parseLatencyRoutes(%q): got %d entries, want %d", tt.input, len(got), len(tt.expected))
			continue
		}
		for k, v := range tt.expected {
			if got[k] != v {
				t.Errorf("parseLatencyRoutes(%q)[%s]: got %v, want %v", tt.input, k, got[k], v)
			}
		}
	}
}
