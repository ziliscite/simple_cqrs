package ports

import "github.com/gin-gonic/gin"

type Handler interface {
	Run(addr string) error
	GetProduct(c *gin.Context)
	SearchProduct(c *gin.Context)
}
