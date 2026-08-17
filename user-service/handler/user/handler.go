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
