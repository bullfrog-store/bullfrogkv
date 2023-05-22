package server

import (
	"bullfrogkv/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Bullfrog request paths
const (
	pathSet    = "/set"
	pathGet    = "/get"
	pathDelete = "/delete"
)

func Router(srv *BullfrogServer) *gin.Engine {
	ginsrv := gin.New()
	ginsrv.GET(pathSet, srv.handlerSet)
	ginsrv.GET(pathGet, srv.handlerGet)
	ginsrv.GET(pathDelete, srv.handlerDelete)
	return ginsrv
}

func (srv *BullfrogServer) handlerSet(c *gin.Context) {
	key, value := c.Query("key"), c.Query("value")

	if err := srv.engine.Set(toBytes(key), toBytes(value)); err != nil {
		logger.Warningf("set error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": msgSetFailure,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msgSetSuccess,
	})
}

func (srv *BullfrogServer) handlerGet(c *gin.Context) {
	key := c.Query("key")

	value, err := srv.engine.Get(toBytes(key))
	if err != nil {
		logger.Warningf("get error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": msgGetFailure,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msgGetSuccess,
		"key":     key,
		"value":   string(value),
	})
}

func (srv *BullfrogServer) handlerDelete(c *gin.Context) {
	key := c.Query("key")

	if err := srv.engine.Delete(toBytes(key)); err != nil {
		logger.Warningf("delete error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": msgDeleteFailure,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": msgDeleteSuccess,
	})
}
