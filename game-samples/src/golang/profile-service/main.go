package main

import (
	"context"
	"log"
	"net/http"

	spanner "cloud.google.com/go/spanner"
	"github.com/dtest/spanner-game-profile-service/models"
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

// TODO: used by authentication server to generate load. Should not be called by other entities,
//  so restrictions should be implemented
func getPlayerUUIDs(c *gin.Context) {
	client, ctx := getSpannerConnection(c)

	players, err := models.GetPlayerUUIDs(ctx, client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No players exist"})
		return
	}

	c.IndentedJSON(http.StatusOK, players)
}

func getPlayerByID(c *gin.Context) {
	var playerUUID = c.Param("id")

	client, ctx := getSpannerConnection(c)

	player, err := models.GetPlayerByUUID(playerUUID, ctx, client)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "player not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, player)
}

func createPlayer(c *gin.Context) {
	var player models.Player

	if err := c.BindJSON(&player); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client, ctx := getSpannerConnection(c)
	playerUUID, err := models.AddPlayer(player, ctx, client)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, playerUUID)
}

func main() {
	router := gin.Default()
	// TODO: Better configuration of trusted proxy
	router.SetTrustedProxies(nil)

	var db = "projects/development-344820/instances/cymbal-games/databases/my_game"
	router.Use(setSpannerConnection(db))

	router.GET("/players", getPlayerUUIDs)
	router.GET("/players/:id", getPlayerByID)
	router.POST("/players", createPlayer)

	// TODO: Better configuration of host
	router.Run("localhost:8080")
}
