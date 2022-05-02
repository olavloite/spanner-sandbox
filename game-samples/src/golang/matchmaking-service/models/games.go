package models

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	spanner "cloud.google.com/go/spanner"
	"github.com/google/uuid"
	iterator "google.golang.org/api/iterator"
)

type Game struct {
	GameUUID string   `json:"gameUUID"`
	Players  []string `json:"players"`
	Winner   string   `json:"winner"`
	created  time.Time
	finished time.Time
}

func generateUUID() string {
	return uuid.NewString()
}

// Read rows from a spanner.
func readRows(iter *spanner.RowIterator) ([]spanner.Row, error) {
	var rows []spanner.Row
	defer iter.Stop()

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return rows, err
		}

		rows = append(rows, *row)
	}

	return rows, nil
}

// Get players for provided game
func getGamePlayers(gameUUID string, ctx context.Context, client spanner.Client) ([]string, error) {
	txn := client.ReadOnlyTransaction()
	stmt := spanner.Statement{
		SQL: `SELECT PlayerUUID FROM games g, UNNEST(g.Players) AS PlayerUUID WHERE gameUUID=@game`,
		Params: map[string]interface{}{
			"game": gameUUID,
		},
	}

	iter := txn.Query(ctx, stmt)

	var playerUUIDs []string

	playerRows, err := readRows(iter)

	if err != nil {
		return playerUUIDs, err
	}

	for _, row := range playerRows {
		var pUUID string
		if err := row.Columns(&pUUID); err != nil {
			return playerUUIDs, err
		}

		playerUUIDs = append(playerUUIDs, pUUID)
	}

	return playerUUIDs, nil
}

// Provided a game UUID, determine the winner
// Current implementation is a random player from the list of players assigned to the game
func determineWinner(playerUUIDs []string) string {
	if len(playerUUIDs) == 0 {
		return ""
	}

	var winnerUUID string

	rand.Seed(time.Now().UnixNano())
	offset := rand.Intn(len(playerUUIDs))
	winnerUUID = playerUUIDs[offset]
	return winnerUUID
}

// Create a new game and assign players
// Players that are not currently playing a game are eligble to be selected for the new game
func CreateGame(g Game, ctx context.Context, client spanner.Client) (string, error) {
	// Initialize game values
	g.GameUUID = generateUUID()

	numPlayers := 100

	// Create and assign
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		// get players
		query := fmt.Sprintf("SELECT playerUUID FROM (SELECT playerUUID FROM players WHERE current_game IS NULL) TABLESAMPLE RESERVOIR (%d ROWS)", numPlayers)
		stmt := spanner.Statement{SQL: query}
		iter := txn.Query(ctx, stmt)

		var playerUUIDs []string

		playerRows, err := readRows(iter)
		if err != nil {
			return err
		}

		for _, row := range playerRows {
			var pUUID string
			if err := row.Columns(&pUUID); err != nil {
				return err
			}

			playerUUIDs = append(playerUUIDs, pUUID)
		}

		// Create the game
		stmt = spanner.Statement{
			SQL: `INSERT games (gameUUID, players, created) VALUES
					(@gameUUID, @players, CURRENT_TIMESTAMP())
			`,
			Params: map[string]interface{}{
				"gameUUID": g.GameUUID,
				"players":  playerUUIDs,
			},
		}
		_, err = txn.Update(ctx, stmt)

		// Update players to lock into this game
		stmt = spanner.Statement{
			SQL: `UPDATE players SET current_game = @game WHERE playerUUID IN UNNEST(@players)`,
			Params: map[string]interface{}{
				"game":    g.GameUUID,
				"players": playerUUIDs,
			},
		}
		_, err = txn.Update(ctx, stmt)

		return err
	})

	if err != nil {
		return "", err
	}

	// return player object if successful, else return nil
	return g.GameUUID, nil
}

// Closing game. When provided a Game, chose a random winner and close out the game.
// A game is closed by setting the winner and finished time.
// Additionally all players' game stats are updated, and the current_game is set to null to allow
// them to be chosen for a new game.
func CloseGame(g Game, ctx context.Context, client spanner.Client) (string, error) {
	// Get game players
	playerUUIDs, err := getGamePlayers(g.GameUUID, ctx, client)

	if err != nil {
		return "", err
	}

	// Might be an issue if there are no players!
	if len(playerUUIDs) == 0 {
		errorMsg := fmt.Sprintf("No players found for game '%s'", g.GameUUID)
		return "", errors.New(errorMsg)
	}

	// Get random winner
	winnerUUID := determineWinner(playerUUIDs)

	// Close game
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `UPDATE games SET finished=CURRENT_TIMESTAMP(), winner=@winner WHERE gameUUID=@game AND finished IS NULL`,
			Params: map[string]interface{}{
				"game":   g.GameUUID,
				"winner": winnerUUID,
			},
		}
		rowCount, err := txn.Update(ctx, stmt)

		if err != nil {
			return err
		}

		// If number of rows updated is not 1, then we have a problem. Don't do anything else
		if rowCount != 1 {
			errorMsg := fmt.Sprintf("Error closing game '%s'", g.GameUUID)
			return errors.New(errorMsg)
		}

		// Update each player to increment stats.games_played (and stats.games_won if winner), and set current_game to null
		// so they can be chosen for a new game

		// TODO: update stats. This requires having a copy of the stats within the transaction that can be modified
		for _, player := range playerUUIDs {
			// Update player
			stmt = spanner.Statement{
				SQL: `UPDATE players SET current_game = NULL WHERE current_game=@game AND playerUUID=@player`,
				Params: map[string]interface{}{
					// "stats": g.Stats,
					"game":   g.GameUUID,
					"player": player,
				},
			}

			// execute transaction10
			_, err = txn.Update(ctx, stmt)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return winnerUUID, nil
}
