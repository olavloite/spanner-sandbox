CREATE TABLE players (
	playerUUID STRING(36) NOT NULL,
	player_name STRING(64) NOT NULL,
    email STRING(MAX) NOT NULL,
	user_password STRING(61) NOT NULL,
	created TIMESTAMP,
	updated TIMESTAMP,
    active_skinUUID STRING(36) NOT NULL,
	stats JSON,
	account_balance NUMERIC,
	is_logged_in BOOL,
	last_login TIMESTAMP,
	valid_email BOOL
) PRIMARY KEY (playerUUID);

CREATE UNIQUE INDEX PlayerAuthentication ON players(email) STORING(user_password);
CREATE UNIQUE INDEX PlayerName ON players(player_name);
