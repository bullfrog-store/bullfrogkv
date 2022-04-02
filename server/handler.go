package main

import (
	"bullfrogkv/storage"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type putReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type deleteReq struct {
	Key string `json:"key"`
}

func (e *raftEngine) putKVHandle(c *gin.Context) {
	var data putReq
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	key := byteForm(data.Key)
	val := byteForm(data.Value)
	err = e.engine.Set(key, val)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": writeFail,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": writeSuccess,
	})
}

func (e *raftEngine) getKVHandle(c *gin.Context) {
	key := c.Query("key")
	val, err := e.engine.Get(byteForm(key))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			log.Println("get error:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": getSuccess,
		"key":     key,
		"data":    string(val),
	})
}

func (e *raftEngine) delKVHandle(c *gin.Context) {
	var data deleteReq
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	key := byteForm(data.Key)
	err = e.engine.Delete(key)
	if err != nil {
		log.Println("del error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": deleteFail,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": deleteSuccess,
	})
}

func byteForm(data string) []byte {
	return []byte(data)
}
