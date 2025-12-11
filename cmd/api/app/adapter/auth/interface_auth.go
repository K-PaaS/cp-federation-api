package auth

import "github.com/gin-gonic/gin"

type AuthAdapter interface {
	AuthMiddleware() gin.HandlerFunc
	ExtractUserID(c *gin.Context) (string, error)
	IsAdminCheck(userID string) (bool, error)
}
