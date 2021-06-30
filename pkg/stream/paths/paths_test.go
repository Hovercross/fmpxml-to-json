package paths_test

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
	"github.com/hovercross/fmpxml-to-json/pkg/stream/paths"
)

func TestSpaceChain_IsExact(t *testing.T) {

	tests := []struct {
		left  paths.SpaceChain
		right paths.SpaceChain
		want  bool
	}{
		{paths.Database,
			paths.SpaceChain{
				xml.Name{Space: constants.SPACE, Local: constants.FMPXMLRESULT},
				xml.Name{Space: constants.SPACE, Local: constants.DATABASE}},
			true,
		},
		{paths.Database,
			paths.SpaceChain{
				xml.Name{Space: constants.SPACE, Local: constants.FMPXMLRESULT},
				xml.Name{Space: constants.SPACE, Local: constants.DATABASE},
				xml.Name{Space: constants.SPACE, Local: constants.FIELD}},
			false,
		},
		{paths.Database,
			paths.SpaceChain{
				xml.Name{Space: constants.SPACE, Local: constants.FMPXMLRESULT},
				xml.Name{Space: constants.SPACE, Local: constants.FIELD}},
			false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			if got := tt.left.IsExact(tt.right); got != tt.want {
				t.Errorf("SpaceChain.IsExact() = %v, want %v", got, tt.want)
			}
		})
	}
}
