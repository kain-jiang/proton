package cerr

import "net/http"

const appCode = 24
const (
	ServerProduceError          = iota + 1_000*appCode + 1_000_000*http.StatusInternalServerError
	ParamsInvalidError          = iota + 1_000*appCode + 1_000_000*http.StatusBadRequest
	PluginNotFoundError         = iota + 1_000*appCode + 1_000_000*http.StatusNotFound
	ComponentAlreadyExistsError = iota + 1_000*appCode + 1_000_000*http.StatusConflict
	ComponentNotFoundError      = iota + 1_000*appCode + 1_000_000*http.StatusNotFound
	originalProduceError        = iota + 1_000*appCode + 0
)

func OriginalProduceErrorWithStatus(statusCode int32) int {
	httpCode := http.StatusInternalServerError
	if statusCode >= 200 && statusCode < 600 {
		httpCode = int(statusCode)
	}
	return originalProduceError + 1_000_000*httpCode
}
