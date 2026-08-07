-- name: GetUserByID :one
SELECT id, name, email, created_at
FROM users
WHERE id = $1;


--name: RegisterUser:one
INSERT INTO users(username,email,password)
VALUES($1,$2,$3) 
RETURNING id,email,username;
