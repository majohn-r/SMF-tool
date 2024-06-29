package commands

import (
	"testing"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
	}{"dummy": {}}
	for name := range tests {
		t.Run(name, func(t *testing.T) {
			Load()
		})
	}
}
