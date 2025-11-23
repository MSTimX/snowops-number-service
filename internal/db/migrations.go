package db

import (
	"fmt"

	"gorm.io/gorm"
)

var migrationStatements = []string{
	`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
	`CREATE TABLE IF NOT EXISTS plates (
		id              BIGSERIAL PRIMARY KEY,
		number          TEXT NOT NULL,
		normalized      TEXT NOT NULL,
		country         TEXT,
		region          TEXT,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS ux_plates_normalized ON plates(normalized);`,
	`CREATE TABLE IF NOT EXISTS lists (
		id          BIGSERIAL PRIMARY KEY,
		name        TEXT NOT NULL,
		type        TEXT NOT NULL,
		description TEXT,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS ux_lists_name ON lists(name);`,
	`CREATE TABLE IF NOT EXISTS list_items (
		list_id     BIGINT REFERENCES lists(id),
		plate_id    BIGINT REFERENCES plates(id),
		note        TEXT,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
		PRIMARY KEY (list_id, plate_id)
	);`,
	`DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM lists WHERE name = 'default_whitelist') THEN
			INSERT INTO lists (name, type, description) VALUES ('default_whitelist', 'WHITELIST', 'Default whitelist');
		END IF;
		IF NOT EXISTS (SELECT 1 FROM lists WHERE name = 'default_blacklist') THEN
			INSERT INTO lists (name, type, description) VALUES ('default_blacklist', 'BLACKLIST', 'Default blacklist');
		END IF;
	END
	$$;`,
}

func runMigrations(db *gorm.DB) error {
	for i, stmt := range migrationStatements {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	return nil
}

