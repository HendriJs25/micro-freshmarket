package role

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	apperror "user-service/common/error"
	"user-service/common/response"
	"user-service/domain/dto/request"
	responsedto "user-service/domain/dto/response"
	roleservice "user-service/services/role"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	roleService roleservice.Service
}

func NewHandler(roleService roleservice.Service) *Handler {
	return &Handler{
		roleService: roleService,
	}
}

func (h *Handler) Create(c *gin.Context) {
	var req request.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.Response{
			Message: "invalid create role request",
			Data:    nil,
		})
		return
	}

	err := h.roleService.Create(c.Request.Context(), roleservice.CreateInput{
		Name: req.Name,
	})

	if err != nil {
		h.respondCreateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Message: "Success",
		Data:    nil,
	})
}

func (h *Handler) GetAll(c *gin.Context) {
	search := c.Query("search")

	roles, err := h.roleService.GetAll(c.Request.Context(), search)
	if err != nil {
		log.Printf("get roles failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
		return
	}

	roleResponses := make([]responsedto.RoleResponse, 0, len(roles))

	for _, role := range roles {
		roleResponses = append(roleResponses, responsedto.RoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	c.JSON(http.StatusOK, response.Response{
		Message: "success",
		Data:    roleResponses,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	roleIDString := strings.TrimSpace(c.Param("id"))

	roleID, err := strconv.ParseInt(roleIDString, 10, 64)
	if err != nil || roleID <= 0 {
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "missing or invalid role id",
			Data:    nil,
		})
		return
	}

	role, err := h.roleService.GetByID(c.Request.Context(), roleID)
	if err != nil {
		h.respondGetByIDError(c, err)
		return
	}

	roleResponse := responsedto.RoleResponse{
		ID:   role.ID,
		Name: role.Name,
	}

	c.JSON(http.StatusOK, response.Response{
		Message: "success",
		Data:    roleResponse,
	})
}

func (h *Handler) respondGetByIDError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "missing or invalid role id",
			Data:    nil,
		})
	case errors.Is(err, apperror.ErrNotFound):
		c.JSON(http.StatusNotFound, response.Response{
			Message: "role not found",
			Data:    nil,
		})
	default:
		log.Printf("get role by id failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})

	}
}

func (h *Handler) respondCreateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidArgument):
		c.JSON(http.StatusUnprocessableEntity, response.Response{
			Message: "invalid create role request",
			Data:    nil,
		})
	case errors.Is(err, apperror.ErrAlreadyExists):
		c.JSON(http.StatusConflict, response.Response{
			Message: "role already exists",
			Data:    nil,
		})
	default:
		log.Printf("create role failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
	}
}
