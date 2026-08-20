package middleware

import (
	"net/http"
	"strings"
	"user-service/common/response"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowedRoleSet := make(map[string]struct{}, len(allowedRoles))

	for _, role := range allowedRoles {
		role = strings.TrimSpace(role)

		if role == "" {
			continue
		}

		allowedRoleSet[role] = struct{}{}
	}

	return func(c *gin.Context) {
		identity, ok := IdentityFromContext(c)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
				Message: "missing or invalid token",
				Data:    nil,
			})
			return
		}

		_, allowed := allowedRoleSet[identity.RoleName]

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Response{
				Message: "insufficient permissions",
				Data:    nil,
			})
			return
		}
		c.Next()
	}
}
