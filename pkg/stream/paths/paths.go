package paths

import (
	"github.com/hovercross/fmpxml-to-json/pkg/stream/constants"
)

var (
	ErrorCode = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.ERRORCODE)
	Product   = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.PRODUCT)
	Database  = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.DATABASE)
	Field     = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.METADATA, constants.FIELD)
	Row       = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.RESULTSET, constants.ROW)
	Col       = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.RESULTSET, constants.ROW, constants.COL)
	Data      = makeSpaceChain(constants.SPACE, constants.FMPXMLRESULT, constants.RESULTSET, constants.ROW, constants.COL, constants.DATA)
)
