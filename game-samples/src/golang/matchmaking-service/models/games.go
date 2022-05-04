package models

import (
	"context"
	"encoding/json"
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

type PlayerStats struct {
	Games_played int `json:"games_played"`
	Games_won    int `json:"games_won"`
}

type Player struct {
	PlayerUUID string           `json:"playerUUID"`
	Stats      spanner.NullJSON `json:"stats"`
}

func generateUUID() string {
	return uuid.NewString()
}

// Helper function to read rows from a spanner.
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

// Get players for a game
// We only care about the playerUUID and their stats, as this is intended to be used
// to modify players when a game is closed
func getGamePlayers(gameUUID string, ctx context.Context, txn *spanner.ReadWriteTransaction) ([]string, []Player, error) {
	stmt := spanner.Statement{
		SQL: `SELECT PlayerUUID, Stats FROM players
				INNER JOIN (
				SELECT pUUID FROM games g, UNNEST(g.Players) AS pUUID WHERE gameUUID=@game
				) AS gPlayers ON gPlayers.pUUID = players.PlayerUUID;`,
		Params: map[string]interface{}{
			"game": gameUUID,
		},
	}

	iter := txn.Query(ctx, stmt)
	playerRows, err := readRows(iter)
	if err != nil {
		return []string{}, []Player{}, err
	}

	var playerUUIDs []string
	var players []Player
	for _, row := range playerRows {
		var p Player

		if err := row.ToStruct(&p); err != nil {
			return []string{}, []Player{}, err
		}
		if p.Stats.IsNull() {
			// Initialize player stats
			p.Stats = spanner.NullJSON{Value: PlayerStats{
				Games_played: 0,
				Games_won:    0,
			}, Valid: true}
		}

		players = append(players, p)
		playerUUIDs = append(playerUUIDs, p.PlayerUUID)
	}

	return playerUUIDs, players, nil
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

// Given a list of players and a winner's UUID, update players of a game
// Updating players involves closing out the game (current_game = NULL) and
// updating their game stats. Specifically, we are incrementing games_played.
// If the player is the determined winner, then their games_won stat is incremented.
func updateGamePlayers(players []Player, winnerUUID string, gameUUID string,
	ctx context.Context, txn *spanner.ReadWriteTransaction) error {
	for _, p := range players {
		// Modify stats
		var pStats PlayerStats
		json.Unmarshal([]byte(p.Stats.String()), &pStats)

		pStats.Games_played = pStats.Games_played + 1

		if p.PlayerUUID == winnerUUID {
			pStats.Games_won = pStats.Games_won + 1
		}
		updatedStats, _ := json.Marshal(pStats)
		p.Stats.UnmarshalJSON(updatedStats)

		// Update player
		stmt := spanner.Statement{
			SQL: `UPDATE players SET current_game = NULL, stats=@pStats WHERE current_game=@game AND playerUUID=@player`,
			Params: map[string]interface{}{
				// "stats": g.Stats,
				"game":   gameUUID,
				"pStats": p.Stats,
				"player": p.PlayerUUID,
			},
		}

		// execute transaction
		_, err := txn.Update(ctx, stmt)

		if err != nil {
			return err
		}
	}

	return nil
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

		playerRows, err := readRows(iter)
		if err != nil {
			return err
		}

		var playerUUIDs []string

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
	// Close game
	var winnerUUID string
	_, err := client.ReadWriteTransaction(ctx,
		func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
			// Get game players
			playerUUIDs, players, err := getGamePlayers(g.GameUUID, ctx, txn)

			if err != nil {
				return err
			}

			// Might be an issue if there are no players!
			if len(playerUUIDs) == 0 {
				errorMsg := fmt.Sprintf("No players found for game '%s'", g.GameUUID)
				return errors.New(errorMsg)
			}

			// Get random winner
			winnerUUID = determineWinner(playerUUIDs)

			stmt := spanner.Statement{
				SQL: `UPDATE games SET finished=CURRENT_TIMESTAMP(), winner=@winner WHERE gameUUID=@game AND finished IS NULL`,
				Params: map[string]interface{}{
					"game":   g.GameUUID,
					"winner": winnerUUID,
				},
			}
			rowCount, gameErr := txn.Update(ctx, stmt)

			if gameErr != nil {
				return gameErr
			}

			// If number of rows updated is not 1, then we have a problem. Don't do anything else
			if rowCount != 1 {
				errorMsg := fmt.Sprintf("Error closing game '%s'", g.GameUUID)
				return errors.New(errorMsg)
			}

			// Update each player to increment stats.games_played (and stats.games_won if winner),
			// and set current_game to null so they can be chosen for a new game
			playerErr := updateGamePlayers(players, winnerUUID, g.GameUUID, ctx, txn)
			if playerErr != nil {
				return playerErr
			}

			return nil
		})

	if err != nil {
		return "", err
	}

	return winnerUUID, nil
}
