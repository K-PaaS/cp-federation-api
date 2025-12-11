package response

import (
	"errors"
	"github.com/gin-gonic/gin"
	apperrors "github.com/karmada-io/dashboard/cmd/api/app/errors"
	"github.com/karmada-io/dashboard/cmd/api/app/localize"
	"github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	errmsg "github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"k8s.io/klog/v2"
	"net/http"
)

// BaseResponse is the base response
type BaseResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`
}

// Success generate success response
func Success(c *gin.Context, obj interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, obj)
}

func Created(c *gin.Context) {
	code := http.StatusCreated
	c.AbortWithStatusJSON(code, BaseResponse{
		Code: code,
		Msg:  localize.GetLocalizeMessage(c, msgkey.ResourceCreateSuccess),
	})
}

func SuccessWithMessage(c *gin.Context, msg string) {
	code := http.StatusOK
	c.AbortWithStatusJSON(code, BaseResponse{
		Code: code,
		Msg:  localize.GetLocalizeMessage(c, msg),
	})
}

func Unauthorized(c *gin.Context, msg string) {
	code := http.StatusUnauthorized
	c.AbortWithStatusJSON(code, BaseResponse{
		Code: code,
		Msg:  localize.GetLocalizeMessage(c, msg),
	})
}

func ServerError(c *gin.Context) {
	Failed(c, errmsg.RequestFailed)
}

func FailedWithError(c *gin.Context, err error) {
	klog.ErrorS(err, "handling error")

	var httpErr *apperrors.HttpError
	if errors.As(err, &httpErr) {
		Failed(c, httpErr.Msg, httpErr.Code)
		return
	}

	// fallback: 예상치 못한 에러
	Failed(c, errmsg.RequestFailed)
}

func Failed(c *gin.Context, msg string, statusCode ...int) {
	code := http.StatusInternalServerError
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	c.AbortWithStatusJSON(code, BaseResponse{
		Code: code,
		Msg:  localize.GetLocalizeMessage(c, msg),
	})
}
