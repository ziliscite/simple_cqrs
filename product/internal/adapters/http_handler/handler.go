package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/cqrs_product/internal/application"
	"github.com/ziliscite/cqrs_product/internal/application/command"
	"github.com/ziliscite/cqrs_product/internal/ports"
	"net/http"
)

type handler struct {
	app application.Service
	en  *gin.Engine
}

func NewHandler(app application.Service) ports.Handler {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	return &handler{
		app: app,
		en:  r,
	}
}

func (h *handler) Run(addr string) error {
	h.setupRoutes()
	return h.en.Run(addr)
}

func (h *handler) setupRoutes() {
	h.en.POST("/products", h.CreateProduct)
	h.en.PATCH("/products/:id", h.UpdateProduct)
	h.en.DELETE("/products/:id", h.DeleteProduct)

	// health check
	h.en.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (h *handler) CreateProduct(c *gin.Context) {
	var request struct {
		Name     string  `json:"name"`
		Category string  `json:"category"`
		Price    float64 `json:"price"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
		return
	}

	cmd, errs := command.NewCreateProduct(request.Name, request.Category, request.Price)
	if errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
		return
	}

	if err := h.app.Create.Handle(c, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *handler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id param"})
		return
	}

	var request struct {
		Name     string  `json:"name"`
		Category string  `json:"category"`
		Price    float64 `json:"price"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
		return
	}

	cmd, errs := command.NewUpdateProduct(id, request.Name, request.Category, request.Price)
	if errs != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
		return
	}

	if err := h.app.Update.Handle(c, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *handler) DeleteProduct(c *gin.Context) {
	// get id from the path param
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id param"})
		return
	}

	cmd, err := command.NewDeleteProduct(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err = h.app.Delete.Handle(c, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
