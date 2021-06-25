package timeconv_test

import (
	"fmt"
	"testing"

	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

func TestParseDateFormat(t *testing.T) {
	tests := []struct {
		fmt  string
		want string
	}{
		{"MM/dd/yy", "01/02/06"},
		{"M/d/yyyy", "1/2/2006"},
		{"yyyy-mm-dd", "2006-01-02"},
		{"M-d-yyyy", "1-2-2006"},
		{"Mdyyyy", "122006"}, // This will, pretty much unambiguously, fail to parse - but our algorithm allows it. Don't do this.
		{"M-d/yyyy", "1-2/2006"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			if got := timeconv.ParseDateFormat(tt.fmt); got != tt.want {
				t.Errorf("ParseDateFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
