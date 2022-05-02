CREATE TABLE games (
	gameUUID string(36) NOT NULL,
	players ARRAY<STRING(36)> NOT NULL,
    winner STRING(36),
	created TIMESTAMP,
    finished TIMESTAMP
) PRIMARY KEY (gameUUID);

ALTER TABLE players ADD COLUMN current_game STRING(36);
