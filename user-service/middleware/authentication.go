package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"
	apperror "user-service/common/error"
	"user-service/common/response"
	sessionrepository "user-service/repository/session"
	jwtservice "user-service/services/jwt"

	"github.com/gin-gonic/gin"
)

const authenticatedIdentityKey = "authenticated_identity"

type Identity struct {
	UserID   int64
	Name     string
	Email    string
	RoleName string
}

type Authentication struct {
	jwtService        jwtservice.Service
	sessionRepository sessionrepository.Repository
}

func NewAuthentication(jwtService jwtservice.Service, sessionRepository sessionrepository.Repository) *Authentication {
	return &Authentication{
		jwtService:        jwtService,
		sessionRepository: sessionRepository,
	}
}

func (a *Authentication) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))

		if err != nil {
			abortUnauthorized(c)
			return
		}

		claims, err := a.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		session, err := a.sessionRepository.Get(c.Request.Context(), tokenString)
		if err != nil {
			if errors.Is(err, apperror.ErrNotFound) {
				abortUnauthorized(c)
				return
			}

			log.Printf("authentication session lookup failed: %v", err)
			abortUnauthorized(c)
			return
		}

		if session.UserID != claims.UserID {
			log.Printf("authentication identity mismatch jwt_user_id=%d session_user_id=%d", claims.UserID, session.UserID)
			abortUnauthorized(c)
			return
		}

		if strings.TrimSpace(session.RoleName) == "" {
			log.Printf("authentication session has empty role user_id=%d", session.UserID)
			abortInternalServerError(c)
			return
		}

		c.Set(authenticatedIdentityKey, Identity{
			UserID:   session.UserID,
			Name:     session.Name,
			Email:    session.Email,
			RoleName: session.RoleName,
		})

		c.Next()
	}
}

func IdentityFromContext(ctx *gin.Context) (Identity, bool) {
	value, exists := ctx.Get(authenticatedIdentityKey)
	if !exists {
		return Identity{}, false
	}

	identity, ok := value.(Identity)
	if !ok {
		return Identity{}, false
	}

	return identity, true
}

func extractBearerToken(authorizationHeader string) (string, error) {
	parts := strings.Fields(authorizationHeader)

	if len(parts) != 2 {
		return "", apperror.ErrInvalidToken
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", apperror.ErrInvalidToken
	}

	tokenString := strings.TrimSpace(parts[1])

	if tokenString == "" {
		return "", apperror.ErrInvalidToken
	}

	return tokenString, nil
}

func abortUnauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")

	c.AbortWithStatusJSON(http.StatusUnauthorized, response.Response{
		Message: "missing or invalid token",
		Data:    nil,
	})
}

func abortInternalServerError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
		Message: "internal server error",
		Data:    nil,
	})
}
