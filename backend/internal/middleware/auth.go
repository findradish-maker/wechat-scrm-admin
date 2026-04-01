package middleware

import (
	"net/http"
	"strings"
	"wechat-enterprise-backend/internal/response"
	"wechat-enterprise-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const userContextKey = "authClaims"

func AuthRequired(manager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, 40101, "未登录或 token 缺失")
			c.Abort()
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := manager.Parse(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, 40102, "token 无效")
			c.Abort()
			return
		}
		c.Set(userContextKey, claims)
		c.Next()
	}
}

func CurrentClaims(c *gin.Context) (*jwt.Claims, bool) {
	value, ok := c.Get(userContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*jwt.Claims)
	return claims, ok
}
