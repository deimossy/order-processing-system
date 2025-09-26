package postgres

const (
	saveUserQuery = `
	INSERT INTO users (
		email, password_hash
	) VALUES (
		:email, :password_hash
	);
    `
	getUserByEmailQuery = `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE email = $1;
	`
	getUserByIDQuery = `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE id = $1;
	`
	deleteUserByIDQuery = `
	DELETE FROM users
	WHERE id = $1;
	`
	saveRefreshTokenQuery = `
	INSERT INTO refresh_tokens (
		user_id, token_hash, expires_at
	) VALUES (
		:user_id, :token_hash, :expires_at
	);
	`
	getRefreshTokenByTokenHashQuery = `
	SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
	FROM refresh_tokens
	WHERE token_hash = $1;
	`
	revokeRefreshTokenByTokenHashQuery = `
	UPDATE refresh_tokens
	SET revoked_at = $1
	WHERE token_hash = $2;
	`
)
