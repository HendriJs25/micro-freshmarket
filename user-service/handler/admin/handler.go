package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Check(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
