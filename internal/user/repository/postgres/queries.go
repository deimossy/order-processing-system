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
	SELECT id, email, password_hash, created_at FROM users WHERE email = $1;
	`
	getUserByIDQuery = `
	SELECT * FROM users WHERE id = $1;
	`
	deleteUserByIDQuery = `
	DELETE FROM users WHERE id = $1;
	`
)
