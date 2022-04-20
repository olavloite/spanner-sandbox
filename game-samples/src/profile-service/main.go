package main

import (
	"context"
	"log"
	"net/http"

	spanner "cloud.google.com/go/spanner"
	"github.com/dtest/spanner-profile-service/models"
	"github.com/gin-gonic/gin"
)

func setSpannerConnection(connectionString string) gin.HandlerFunc {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, connectionString)

	if err != nil {
		log.Fatal(err)
	}

	return func(c *gin.Context) {
		c.Set("spanner_client", *client)
		c.Set("ctx", ctx)
		c.Next()
	}
}

func getSpannerConnection(c *gin.Context) spanner.Client {
	return c.MustGet("spanner_client").(spanner.Client)
}

func getPlayers(c *gin.Context) {
	c.IndentedJSON(http.StatusNotFound, "Page not found")
}

func createPlayer(c *gin.Context) {
	var player models.Player

	if err := c.BindJSON(&player); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := getSpannerConnection(c)
	ctx := c.MustGet("ctx").(context.Context)
	playerUUID, err := models.AddPlayer(player, ctx, client)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, playerUUID)
}

func main() {
	router := gin.Default()

	var db = "projects/development-344820/instances/cymbal-games/databases/my_game"
	router.Use(setSpannerConnection(db))

	router.GET("/players", getPlayers)
	router.POST("/players", createPlayer)

	router.Run("localhost:8080")
}
