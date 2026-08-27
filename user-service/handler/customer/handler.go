package customer

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	apperror "user-service/common/error"
	"user-service/common/response"
	responsedto "user-service/domain/dto/response"
	customerservice "user-service/services/customer"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	customerService customerservice.Service
}

func NewHandler(customerService customerservice.Service) *Handler {
	return &Handler{
		customerService: customerService,
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	page := parseOptionalPositiveInt64(c.Query("page"))
	limit := parseOptionalPositiveInt64(c.Query("limit"))

	result, err := h.customerService.GetAll(c.Request.Context(), customerservice.ListInput{
		Search:    c.Query("search"),
		Page:      page,
		Limit:     limit,
		OrderBy:   c.Query("order_by"),
		OrderType: c.Query("order_type"),
	})

	if err != nil {
		h.respondGetAllError(c, err)
		return
	}

	customers := make([]responsedto.CustomerListResponse, 0, len(result.Customers))

	for _, customer := range result.Customers {
		customers = append(customers, responsedto.CustomerListResponse{
			ID:    customer.ID,
			Name:  customer.Name,
			Email: customer.Email,
			Phone: customer.Phone,
			Photo: customer.Photo,
		})
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Message: "Data retrieved successfully",
		Data:    customers,
		Pagination: &response.Pagination{
			Page:       result.Pagination.Page,
			TotalCount: result.Pagination.TotalCount,
			PerPage:    result.Pagination.PerPage,
			TotalPage:  result.Pagination.TotalPage,
		},
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	customerIDString := strings.TrimSpace(c.Param("id"))

	customerID, err := strconv.ParseInt(customerIDString, 10, 64)
	if err != nil || customerID <= 0 {
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "missing or invalid customer ID",
			Data:    nil,
		})
		return
	}

	customer, err := h.customerService.GetByID(c.Request.Context(), customerID)
	if err != nil {
		h.respondGetByID(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Message: "success get customer by id",
		Data: responsedto.CustomerResponse{
			RoleID:  customer.RoleID,
			ID:      customer.ID,
			Name:    customer.Name,
			Email:   customer.Email,
			Phone:   customer.Phone,
			Lat:     customer.Lat,
			Lng:     customer.Lng,
			Address: customer.Address,
			Photo:   customer.Photo,
		},
	})
}

func parseOptionalPositiveInt64(value string) int64 {
	value = strings.TrimSpace(value)

	if value == "" {
		return 0
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return result
}

func (h *Handler) respondGetAllError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "invalid customer query",
			Data:    nil,
		})
	default:
		log.Printf("get customer failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
	}
}

func (h *Handler) respondGetByID(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, response.Response{
			Message: "missing or invalid customer ID",
			Data:    nil,
		})
	case errors.Is(err, apperror.ErrNotFound):
		c.JSON(http.StatusNotFound, response.Response{
			Message: "customer not found",
			Data:    nil,
		})
	default:
		log.Printf("get customer by id failed: %v", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Message: "internal server error",
			Data:    nil,
		})
	}
}
