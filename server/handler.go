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

func (k *kvEngine) putKVHandle(c *gin.Context) {
	var data putReq
	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	key := getBytes(data.Key)
	val := getBytes(data.Value)
	err = k.kve.WriteKV(storage.Modify{
		Data: storage.Put{
			Key:   key,
			Value: val,
			Sync:  true,
		},
	})
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

func (k *kvEngine) getKVHandle(c *gin.Context) {
	keyI := c.Query("key")
	key := getBytes(keyI)
	val, err := k.kve.ReadKV(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		} else {
			log.Println("get error:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": getFail,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": getSuccess,
		"key":     keyI,
		"data":    string(val),
	})
}

func (k *kvEngine) delKVHandle(c *gin.Context) {
	key := c.Query("key")
	err := k.kve.WriteKV(storage.Modify{
		Data: storage.Delete{
			Key:  getBytes(key),
			Sync: true,
		},
	})

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

func getBytes(data string) []byte {
	return []byte(data)
}
