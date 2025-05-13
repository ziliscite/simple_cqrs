package ports

import "github.com/gin-gonic/gin"

type Handler interface {
	Run(addr string) error
	CreateProduct(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
}
