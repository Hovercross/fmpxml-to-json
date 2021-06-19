package timeconv_test

import (
	"fmt"
	"testing"

	"github.com/hovercross/fmpxml-to-json/pkg/timeconv"
)

func TestDateSplitWithSep(t *testing.T) {
	type args struct {
		filemaker string
		sep       string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		{args{"MM/dd/yy", "/"}, "01/02/06", false},
		{args{"M/d/yyyy", "/"}, "1/2/2006", false},
		{args{"M/d/yyyy", "-"}, "", true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			got, err := timeconv.DateSplitWithSep(tt.args.filemaker, tt.args.sep)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateSplitWithSep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DateSplitWithSep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDateFormat(t *testing.T) {
	tests := []struct {
		fmt  string
		want string
	}{
		{"MM/dd/yy", "01/02/06"},
		{"M/d/yyyy", "1/2/2006"},
		{"yyyy-mm-dd", "2006-01-02"},
		{"M-d-yyyy", "1-2-2006"},
		{"Mdyyyy", ""},
		{"M-d/yyyy", ""},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			if got := timeconv.ParseDateFormat(tt.fmt); got != tt.want {
				t.Errorf("ParseDateFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
