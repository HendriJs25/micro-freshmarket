package user

import (
	"errors"
	"log"
	"net/http"
	apperror "user-service/common/error"
	"user-service/common/response"
	"user-service/domain/dto/request"
	userservice "user-service/services/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService userservice.Service
}

func NewHandler(userService userservice.Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) SignUp(c *gin.Context) {
	var req request.SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.Response{
			Message: "invalid signup request",
			Data:    nil,
		})
		return
	}

	err := h.userService.SignUp(c.Request.Context(),
		userservice.SignUpInput{
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
		})
	if err != nil {
		h.respondSignUpError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Message: "user created successfully",
		Data:    nil,
	})

}

func (h *Handler) VerifyAccount(c *gin.Context) {
	var query request.VerifyAccountQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusUnauthorized, response.Response{
			Message: "missing or invalid token",
			Data:    nil,
		})
		return
	}

	err := h.userService.VerifyAccount(c.Request.Context(), query.Token)
	if err != nil {
		h.respondVerifyAccountError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Message: "success",
		Data:    nil,
	})
}

func (h *Handler) respondSignUpError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidArgument):
		c.JSON(http.StatusUnprocessableEntity, response.Response{
			Message: "invalid signup request",
			Data:    nil,
		})
	case errors.Is(err, apperror.ErrAlreadyExists):
		c.JSON(http.StatusConflict, response.Response{
			Message: "email already registered",
			Data:    nil,
		})
	default:
		log.Printf("signup failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
	}
}

func (h *Handler) respondVerifyAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, response.Response{
			Message: "Token expired or invalid",
			Data:    nil,
		})
	case errors.Is(err, apperror.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, response.Response{
			Message: "Token expired or invalid",
			Data:    nil,
		})
	default:
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
	}
}
