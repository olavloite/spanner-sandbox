package models

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	spanner "cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Timestamp time.Time

type PlayerStats struct {
	time_played int
	stat2       int
}

type Player struct {
	PlayerUUID      string `json:"playerUUID"`
	Player_name     string `json:"player_name" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	created         Timestamp
	updated         Timestamp
	Active_skinUUID string      `json:"active_skinUUID"`
	Stats           PlayerStats `json:"stats"`
	Account_balance float64     `json:"account_balance"`
	last_login      Timestamp
	is_logged_in    bool
	valid_email     bool
}

// TODO check for valid domains, and not allow local domains?
func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// TODO complexity validation
func hashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func validatePassword(pwd string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
}

func generateUUID() string {
	return uuid.NewString()
}

func AddPlayer(p Player, ctx context.Context, client spanner.Client) (string, error) {
	// Ensure email is valid
	if err := validateEmail(p.Email); err != true {
		return "", fmt.Errorf("New player has invalid email '%s'", p.Email)
	}

	// take supplied password+salt, hash. Store in user_password
	newPass, err := hashPassword(p.Password)

	if err != nil {
		return "", errors.New("Unable to hash password")
	}

	p.Password = newPass

	// Generate UUIDv4
	p.PlayerUUID = generateUUID()

	// insert into spanner
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT players (playerUUID, player_name, email, user_password, created, active_skinUUID) VALUES
					(@playerUUID, @playerName, @email, @password, CURRENT_TIMESTAMP(), '1')
			`,
			Params: map[string]interface{}{
				"playerUUID": p.PlayerUUID,
				"playerName": p.Player_name,
				"email":      p.Email,
				"password":   p.Password,
			},
		}

		_, err := txn.Update(ctx, stmt)
		return err
	})

	// todo: Handle 'AlreadyExists' errors
	if err != nil {
		return "", err
	}

	// return player object if successful, else return nil
	return p.PlayerUUID, nil
}

// func GetPlayerByLogin(name string, password string) (Player, error) {

// }

// func GetPlayerByUUID(uuid string) (Player, error) {

// }
