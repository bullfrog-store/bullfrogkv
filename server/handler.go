package main

import (
	"bullfrogkv/storage"
	"errors"
	"fmt"
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
			Sync:  false,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "write fail",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "write success",
	})
}

func (k *kvEngine) getKVHandle(c *gin.Context) {
	keyI := c.Query("key")
	key := getBytes(keyI)
	val, err := k.kve.GetKV(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		} else {
			log.Println("get error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"err": "get value fail"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"mgs":  "get success",
		"key":  keyI,
		"data": string(val),
	})
}

func (k *kvEngine) delKVHandle(c *gin.Context) {
	key := c.Query("key")
	err := k.kve.DelKV(getBytes(key), false)
	if err != nil {
		log.Println("del error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": "del fail"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "del success"})
}

func getBytes(data string) []byte {
	return []byte(data)
}
