package timeconv_test

import (
	"fmt"
	"testing"

	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

func TestParseTimeFormat(t *testing.T) {
	tests := []struct {
		fmt  string
		want string
	}{
		{"hh:mm:ss a", "03:04:05 PM"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			if got := timeconv.ParseTimeFormat(tt.fmt); got != tt.want {
				t.Errorf("ParseTimeFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
