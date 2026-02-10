package oauth

import (
	"fmt"
)

// ErrRawHTTP return http error from other server, it use trait.WrrapperError to store response overview
var ErrRawHTTP = fmt.Errorf("")
