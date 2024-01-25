package test

import (
	"jinx/pkg/util"
	"testing"
)

func TestIsLocalHostOrIP(t *testing.T) {

	tests := []struct {
		input  string
		expect bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"192.168.1.1", false},
		{"178.134.3.2", false},
		{"345.567.557.555", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := util.IsLocalhostOrIP(test.input)
			if result != test.expect {
				t.Errorf("expected %v got %v", test.expect, result)
			}
		})
	}
}
