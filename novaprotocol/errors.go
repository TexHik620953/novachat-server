package novaprotocol

import "fmt"

var (
	ErrorFrameZeroLength     = fmt.Errorf("invalid frame: zero length")
	ErrorFrameSizeMissmatch  = fmt.Errorf("invalid frame: size missmatch")
	ErrorFrameTooLarge       = fmt.Errorf("invalid frame: too large")
	ErrorFrameNoHeader       = fmt.Errorf("invalid frame: no header")
	ErrorFrameInvalidHashSum = fmt.Errorf("invalid frame: hashsum mismatch")
)
