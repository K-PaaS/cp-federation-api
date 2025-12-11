package auth

import (
	"github.com/gin-gonic/gin"
)

type MockAuthAdapter struct{}

func NewMockAuthAdapter() AuthAdapter {
	return &MockAuthAdapter{}
}
func (m MockAuthAdapter) AuthMiddleware() gin.HandlerFunc {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthAdapter) ExtractUserID(c *gin.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthAdapter) IsAdminCheck(userID string) (bool, error) {
	//TODO implement me
	panic("implement me")
}
