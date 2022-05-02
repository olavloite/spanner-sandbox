package main

import (
	"context"
	"log"
	"net/http"

	spanner "cloud.google.com/go/spanner"
	"github.com/dtest/spanner-game-match-service/models"
	"github.com/gin-gonic/gin"
)

// Mutator to create spanner context and client, and set them in gin
func setSpannerConnection(connectionString string) gin.HandlerFunc {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, connectionString)

	if err != nil {
		log.Fatal(err)
	}

	return func(c *gin.Context) {
		c.Set("spanner_client", *client)
		c.Set("spanner_context", ctx)
		c.Next()
	}
}

// Helper function to retrieve spanner client and context
func getSpannerConnection(c *gin.Context) (spanner.Client, context.Context) {
	return c.MustGet("spanner_client").(spanner.Client),
		c.MustGet("spanner_context").(context.Context)
}

// Creating a game assigns a list of players not currently playing a game
func createGame(c *gin.Context) {
	var game models.Game

	client, ctx := getSpannerConnection(c)
	gameID, err := models.CreateGame(game, ctx, client)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, gameID)
}

func closeGame(c *gin.Context) {
	var game models.Game

	if err := c.BindJSON(&game); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client, ctx := getSpannerConnection(c)
	playerUUID, err := models.CloseGame(game, ctx, client)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusOK, playerUUID)
}

func main() {
	router := gin.Default()
	// TODO: Better configuration of trusted proxy
	router.SetTrustedProxies(nil)

	var db = "projects/development-344820/instances/cymbal-games/databases/my_game"
	router.Use(setSpannerConnection(db))

	router.POST("/games/create", createGame)
	router.PUT("/games/close", closeGame)

	// TODO: Better configuration of host
	router.Run("localhost:8081")
}
