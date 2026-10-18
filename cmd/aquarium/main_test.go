package main

import "testing"

func TestRepeat(t *testing.T) {
	tests := []struct {
		args   []string
		random bool
		want   string
	}{
		{nil, true, "aquarium -seed 7"},
		{[]string{"-fish", "5"}, true, "aquarium -fish 5 -seed 7"},
		{[]string{"-seed", "7"}, false, "aquarium -seed 7"},
		{[]string{"-seed", "0"}, true, "aquarium -seed 0 -seed 7"},
	}
	for _, tt := range tests {
		if got := repeat(tt.args, 7, tt.random); got != tt.want {
			t.Errorf("repeat(%q, %v) = %q, want %q", tt.args, tt.random, got, tt.want)
		}
	}
}
