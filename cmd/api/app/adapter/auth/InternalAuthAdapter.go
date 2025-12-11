package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/karmada-io/dashboard/cmd/api/app/intra"
	errmsg "github.com/karmada-io/dashboard/cmd/api/app/msgkey"
	"github.com/karmada-io/dashboard/cmd/api/app/response"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"k8s.io/klog/v2"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type InternalAuthAdapter struct {
}

func NewInternalAuthAdapter() AuthAdapter {
	return &InternalAuthAdapter{}
}

func (i InternalAuthAdapter) IsAdminCheck(userID string) (bool, error) {
	var resp intra.AdminCheck
	sendUrl := strings.Replace(intra.Env.IsSuperAdminCheckUrl, "{userID}", userID, -1)
	err := intra.ApiCall(intra.CommonApi, http.MethodGet, sendUrl, nil, &resp)
	if err != nil {
		klog.ErrorS(err, "AdminCheck ApiCall failed")
		return false, err
	}
	return resp.IsSuperAdmin, nil
}

func (i InternalAuthAdapter) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, errmsg.MissingOrMalformedJwt)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(intra.Env.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			slog.Error("Invalid JWT", "err", err.Error())
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Unauthorized(c, errmsg.TokenExpired)
				return
			}
			response.Unauthorized(c, errmsg.TokenFailed)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			slog.Error("Invalid claims")
			response.Unauthorized(c, errmsg.TokenFailed)
			return
		}

		if expRaw, ok := claims["exp"].(float64); ok {
			exp := int64(expRaw)
			if exp < time.Now().Unix() {
				response.Unauthorized(c, errmsg.TokenExpired)
				return
			}
		} else {
			slog.Error("Missing exp claim")
			response.Unauthorized(c, errmsg.TokenFailed)
			return
		}

		userId, ok := claims[intra.ClaimUserIdKey].(string)
		if !ok {
			slog.Error("Missing userId claim")
			response.Unauthorized(c, errmsg.ApiAccessDenied)
			return
		}

		isAdmin, err := i.IsAdminCheck(userId)
		if err != nil {
			slog.Info("Failed to admin check api call", "err", err)
			response.ServerError(c)
			return
		}

		if !isAdmin {
			slog.Info("Not authorized to access this api")
			response.Unauthorized(c, errmsg.ApiAccessDenied)
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func (i InternalAuthAdapter) ExtractUserID(c *gin.Context) (string, error) {
	claims := common.GetClaims(c)
	userID := claims[intra.ClaimUserIdKey].(string)
	return userID, nil
}
