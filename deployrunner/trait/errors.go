package trait

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tec "taskrunner/error/codes"
	tecode "taskrunner/error/codes/trait"

	"github.com/sirupsen/logrus"
)

type InterruptContext struct {
	context.Context
	TimeoutError *Error
}

func (ctx *InterruptContext) Err() error {
	err := ctx.Context.Err()
	if err == context.Canceled {
		if cause := context.Cause(ctx.Context); cause != nil {
			return cause
		}
	}

	if err == context.DeadlineExceeded {
		return ctx.TimeoutError
	} else if err == nil {
		return nil
	}

	return &Error{
		Internal: ECContextEnd,
		Err:      err,
		Detail:   "",
	}
}

func WithCancelCauesContext(ctx context.Context) (context.Context, func(cause *Error)) {
	ctx0, cancel := context.WithCancelCause(ctx)
	return &InterruptContext{Context: ctx0}, func(cause *Error) {
		cancel(cause)
	}
}

func WithTimeoutCauseContext(ctx context.Context, timeout time.Duration, err *Error) (context.Context, context.CancelFunc) {
	ctx0, cancel := context.WithTimeout(ctx, timeout)
	if err == nil {
		err = &Error{
			Internal: ECTimeout,
			Err:      err,
			Detail:   "",
		}
	}
	return &InterruptContext{Context: ctx0, TimeoutError: err}, cancel
}

// Error wrapper internal error for more info
type Error struct {
	// Err support by outside this project
	Err error
	// Detail more custom detail info
	Detail   interface{}
	Internal int
}

func UnwrapError(err error) *Error {
	err0, ok := err.(*Error)
	if !ok {
		return nil
	}
	return err0
}

func (e *Error) ToJson() string {
	bs, rerr := json.Marshal(e)
	if rerr != nil {
		logrus.Warnf("decode custom Error fail: %s", rerr.Error())
	}
	return string(bs)
}

func (e *Error) Error() string {
	err := ""
	if e.Err != nil {
		err = e.Err.Error()
	}
	return fmt.Sprintf("description: %s, err: %s, detail: %#v", tec.ErrorCache.GetCode(e.Internal).Description, err, e.Detail)
}

// IsInternalError is the  internal error
func IsInternalError(get error, want int) bool {
	yes := false
	if get == nil {
		return false
	}
	if err, ok := get.(*Error); ok {
		if err == nil {
			return false
		}
		yes = want == err.Internal
	}
	return yes
}

// const (
// 	ECNULL = iota
// 	// ErrParam the parama error, check version and document
// 	ErrParam

// 	// ErrUniqueKey the unique key conflict
// 	ErrUniqueKey

// 	// ErrAppTypeNoDefined app no suuport
// 	ErrAppTypeNoDefined

// 	// ErrHelmChartAPIVersion chart api version don't support
// 	ErrHelmChartAPIVersion
// 	//ErrJobExecuting repensent job executing status cause error
// 	ErrJobExecuting
// 	//ErrJobCantStop repensent job can't stop
// 	ErrJobCantStop

// 	//ErrConfigNotComfirm config not confirm
// 	ErrConfigNotComfirm

// 	//ErrComponentUsing is using
// 	ErrComponentUsing

// 	//ErrApplicationFile the application file content error
// 	ErrApplicationFile

// 	//ErrConfigValidate the input config is error
// 	ErrConfigValidate

// 	// ErrNotFound internal error for record not found
// 	ErrNotFound
// 	// ErrHelmRepoNoFound repo not found
// 	ErrHelmRepoNoFound
// 	// ErrHelmChartNoFound repo not found
// 	ErrHelmChartNoFound
// 	// ErrSystemNofound system no found
// 	ErrSystemNofound
// 	//ErrComponentVersionLess component version error
// 	ErrComponentVersionLess
// 	// ErrComponentNotFound the component not found in system
// 	ErrComponentNotFound

// 	// ErrComponentDup the component is duplicate
// 	ErrComponentDup
// 	// ErrAPPlicationComponentTortuous the application component is Tortuous
// 	ErrAPPlicationComponentTortuous

// 	// ErrComponentTypeNotDefined the component tpye is not defined
// 	ErrComponentTypeNotDefined

// 	// ErrComponentInstanceRevission the component instance revission is error
// 	ErrComponentInstanceRevission

// 	// ErrComponentDecodeError the component content error, decode error
// 	ErrComponentDecodeError

// 	// ErrJobOwnerError the job owner error
// 	ErrJobOwnerError

// 	// ErrApplicationStillUse still use by other component
// 	ErrApplicationStillUse

// 	// ErrHelmRepoUnknow
// 	ErrHelmRepoUnknow

// 	ECTimeout
// 	ECJobCancel
// 	ECContextEnd
// 	ECHelmChartNotFound
// 	ECHelmReleaseNotFound
// 	ECHelmReleaseForceUpdate
// 	ECHelmRun
// 	ECHelmTimeout
// 	ECHelmK8s
// 	ECHelmK8SORRender
// 	ECTemplate
// 	ECParseChart
// 	ECExit
// 	ECComponentDefined
// 	ECK8sUnknow
// 	ECNetUnknow
// 	ECSQLUnknow
// 	ECHTTPAPIRawError

// 	// ----------------------
// 	// TASK
// 	ECBaseNode

// 	// ----------------------
// 	// engine
// 	ECNoAvailableWorker
// )

const (
	ECNULL                          = tecode.ECNULL
	ErrParam                        = tecode.ErrParam
	ErrUniqueKey                    = tecode.ErrUniqueKey
	ErrAppTypeNoDefined             = tecode.ErrAppTypeNoDefined
	ErrHelmChartAPIVersion          = tecode.ErrHelmChartAPIVersion
	ErrJobExecuting                 = tecode.ErrJobExecuting
	ErrJobCantStop                  = tecode.ErrJobCantStop
	ErrConfigNotComfirm             = tecode.ErrConfigNotComfirm
	ErrComponentUsing               = tecode.ErrComponentUsing
	ErrApplicationFile              = tecode.ErrApplicationFile
	ErrConfigValidate               = tecode.ErrConfigValidate
	ErrNotFound                     = tecode.ErrNotFound
	ErrHelmRepoNoFound              = tecode.ErrHelmRepoNoFound
	ErrHelmChartNoFound             = tecode.ErrHelmChartNoFound
	ErrSystemNofound                = tecode.ErrSystemNofound
	ErrComponentVersionLess         = tecode.ErrComponentVersionLess
	ErrComponentNotFound            = tecode.ErrComponentNotFound
	ErrComponentDup                 = tecode.ErrComponentDup
	ErrAPPlicationComponentTortuous = tecode.ErrAPPlicationComponentTortuous
	ErrComponentTypeNotDefined      = tecode.ErrComponentTypeNotDefined
	ErrComponentInstanceRevission   = tecode.ErrComponentInstanceRevission
	ErrComponentDecodeError         = tecode.ErrComponentDecodeError
	ErrJobOwnerError                = tecode.ErrJobOwnerError
	ErrApplicationStillUse          = tecode.ErrApplicationStillUse
	ErrHelmRepoUnknow               = tecode.ErrHelmRepoUnknow
	ECTimeout                       = tecode.ECTimeout
	ECJobCancel                     = tecode.ECJobCancel
	ECContextEnd                    = tecode.ECContextEnd
	ECHelmChartNotFound             = tecode.ECHelmChartNotFound
	ECHelmReleaseNotFound           = tecode.ECHelmReleaseNotFound
	ECHelmReleaseForceUpdate        = tecode.ECHelmReleaseForceUpdate
	ECHelmRun                       = tecode.ECHelmRun
	ECHelmTimeout                   = tecode.ECHelmTimeout
	ECHelmK8s                       = tecode.ECHelmK8s
	ECHelmK8SORRender               = tecode.ECHelmK8SORRender
	ECTemplate                      = tecode.ECTemplate
	ECParseChart                    = tecode.ECParseChart
	ECExit                          = tecode.ECExit
	ECComponentDefined              = tecode.ECComponentDefined
	ECK8sUnknow                     = tecode.ECK8sUnknow
	ECNetUnknow                     = tecode.ECNetUnknow
	ECSQLUnknow                     = tecode.ECSQLUnknow
	ECHTTPAPIRawError               = tecode.ECHTTPAPIRawError
	ECBaseNode                      = tecode.ECBaseNode
	ECNoAvailableWorker             = tecode.ECNoAvailableWorker
	ECNoAuthorized                  = tecode.ECNoAuthorized
	ECInvalidAuthorized             = tecode.ECInvalidAuthorized
	ErrApplicationNotFound          = tecode.ErrApplicationNotFound
)
