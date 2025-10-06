package auth

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string // hashed password
	IsSuperuser  bool   `json:"is_superuser"`
}

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal("cannot open DB:", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		is_superuser BOOLEAN DEFAULT FALSE
	);
	`
	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatal("cannot create table:", err)
	}

	log.Println("DB initialized")
}

func GetUserByUsername(username string) (*User, error) {
	row := DB.QueryRow("SELECT id, username, password_hash, is_superuser FROM users WHERE username = ?", username)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsSuperuser)
	return user, err
}

func CreateUser(username, hashedPassword string, isSuperuser bool) error {
	_, err := DB.Exec("INSERT INTO users (username, password_hash, is_superuser) VALUES (?, ?, ?)",
		username, hashedPassword, isSuperuser)
	return err
}

func DeleteUser(username string) error {
	_, err := DB.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}

func ListUsers() ([]*User, error) {
	rows, err := DB.Query("SELECT id, username, password_hash, is_superuser FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsSuperuser); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
