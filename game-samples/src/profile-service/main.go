package main

import (
	"net/http"

	"github.com/dtest/spanner-profile-service/models"
	"github.com/gin-gonic/gin"
)

func getPlayers(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, "Page not found")
}

func createPlayer(c *gin.Context) {
	var player models.Player

	if err := c.BindJSON(&player); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	playerUUID, err := models.AddPlayer(player)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, playerUUID)
}

func main() {
	router := gin.Default()
	router.GET("/players", getPlayers)
	router.POST("/players", createPlayer)

	router.Run("localhost:8080")
}
