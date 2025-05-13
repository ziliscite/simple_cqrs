package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ziliscite/cqrs_search/internal/application"
	"github.com/ziliscite/cqrs_search/internal/application/query"
	"github.com/ziliscite/cqrs_search/internal/domain/product"
	"github.com/ziliscite/cqrs_search/internal/ports"
	"net/http"
	"strconv"
)

type handler struct {
	q  *application.Query
	en *gin.Engine
}

func NewHandler(app *application.Query) ports.Handler {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	return &handler{
		q:  app,
		en: r,
	}
}

func (h *handler) Run(addr string) error {
	h.setupRoutes()
	return h.en.Run(addr)
}

func (h *handler) setupRoutes() {
	h.en.GET("/products", h.SearchProduct)
	h.en.GET("/products/:id", h.GetProduct)

	// health check
	h.en.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (h *handler) GetProduct(c *gin.Context) {
	// get id
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id param"})
		return
	}

	// get product
	q := query.NewGetProduct(id)
	p, err := h.q.Get.Handle(c, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *handler) SearchProduct(c *gin.Context) {
	// extract query parameters
	search := h.extractQueryParams(c)

	// get products
	q := query.NewSearchProduct(search)
	products, err := h.q.Search.Handle(c, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *handler) extractQueryParams(c *gin.Context) *product.Search {
	search := product.NewSearch()

	// Extracting query parameters
	name := c.Query("name")
	category := c.Query("category")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	page := c.Query("page")
	pageSize := c.Query("page_size")
	sortField := c.Query("sort_field")
	sortAsc := c.Query("sort_asc")

	// Set parameters using the builder
	if name != "" {
		search.WithName(name)
	}

	if category != "" {
		search.WithCategory(category)
	}

	if minPrice != "" {
		if minPriceFloat, err := strconv.ParseFloat(minPrice, 64); err == nil {
			search.WithMinPrice(minPriceFloat)
		}
	}
	if maxPrice != "" {
		if maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			search.WithMaxPrice(maxPriceFloat)
		}
	}

	if page != "" {
		if pageInt, err := strconv.Atoi(page); err == nil {
			search.WithPage(pageInt)
		}
	}
	if pageSize != "" {
		if pageSizeInt, err := strconv.Atoi(pageSize); err == nil {
			search.WithPageSize(pageSizeInt)
		}
	}

	if sortField != "" {
		search.WithSortField(sortField)
	}
	if sortAsc != "" {
		if asc, err := strconv.ParseBool(sortAsc); err == nil {
			search.WithSortAsc(asc)
		}
	}

	return search
}
