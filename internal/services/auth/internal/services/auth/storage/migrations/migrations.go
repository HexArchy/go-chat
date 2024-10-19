package migrations

import (
	"gorm.io/gorm"
)

func CreateInitialSchema(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			nickname VARCHAR(255) NOT NULL,
			phone_number VARCHAR(20),
			age INT,
			bio TEXT,
			roles TEXT[],
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
		CREATE INDEX IF NOT EXISTS idx_users_nickname ON users (nickname);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);

		CREATE OR REPLACE FUNCTION update_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE OR REPLACE TRIGGER users_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at();

		CREATE OR REPLACE FUNCTION check_user_age()
		RETURNS TRIGGER AS $$
		BEGIN
			IF NEW.age < 16 THEN
				RAISE EXCEPTION 'User must be at least 16 years old';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE OR REPLACE TRIGGER users_check_age
		BEFORE INSERT OR UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION check_user_age();
	`).Error
}

func SetupSchema(db *gorm.DB, schemaName string) error {
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS " + schemaName).Error; err != nil {
		return err
	}

	return db.Exec("SET search_path TO " + schemaName).Error
}
