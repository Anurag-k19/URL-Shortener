package auth

import (
	"context"
	"database/sql"
	"url_shortner/storage"
)

type user struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// upsertUser creates the user on their first sign-in, or updates their email, name and
// last login time if they've signed in before, and returns their users.id.
func upsertUser(ctx context.Context, googleSub, email, name string) (int64, error) {
	var userID int64
	err := storage.DB.QueryRowContext(ctx, `
		INSERT INTO users (google_sub, email, name)
		VALUES ($1, $2, NULLIF($3, ''))
		ON CONFLICT (google_sub) DO UPDATE
		SET email = EXCLUDED.email,
		    name = EXCLUDED.name,
		    last_login_at = NOW()
		RETURNING id`,
		googleSub, email, name,
	).Scan(&userID)
	return userID, err
}

// findUser loads a user by users.id. It returns sql.ErrNoRows if they don't exist.
// It's a variable so tests can replace it without a running database.
var findUser = func(ctx context.Context, userID int64) (user, error) {
	var u user
	var name sql.NullString
	err := storage.DB.QueryRowContext(ctx,
		"SELECT id, email, name FROM users WHERE id = $1", userID,
	).Scan(&u.ID, &u.Email, &name)
	u.Name = name.String
	return u, err
}
