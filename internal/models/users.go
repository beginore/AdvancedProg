package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID             int
	username       string
	Email          string
	HashedPassword string
	Provider       string
	ProviderID     string
	created_at     time.Time
	Role           string
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users 
    (username, email, password_hash, created_at, role) 
    VALUES ($1, $2, $3, CURRENT_TIMESTAMP, 'user')`

	_, err = m.DB.Exec(stmt, username, email, string(hashedPassword))
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT users.user_id, password_hash  FROM users WHERE email = $1"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}

func hashPassword(password string, salt string) (string, error) {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *UserModel) Get(id int) (*User, error) {
	stmt := `SELECT user_id, username, email, password_hash, created_at, role 
             FROM users WHERE user_id = $1`
	row := m.DB.QueryRow(stmt, id)

	u := &User{}
	err := row.Scan(&u.ID, &u.username, &u.Email, &u.HashedPassword, &u.created_at, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return u, nil
}

func (m *UserModel) GetAllUsers() ([]*User, error) {
	stmt := `SELECT user_id, username, email, role FROM users
`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.username, &u.Email, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m *UserModel) UpdatePassword(hashedPassword string, id int) error {
	stmt := "UPDATE users SET password_hash = $1 WHERE user_id = $2"
	_, err := m.DB.Exec(stmt, hashedPassword, id)
	return err
}
func (m *UserModel) GetOrCreateOAuthUser(email, username, provider, provider_id string) (int, error) {
	var userID int

	err := m.DB.QueryRow(`
        SELECT user_id FROM users 
        WHERE provider = $1 AND provider_id = $2`,
		provider, provider_id).Scan(&userID)

	if err == nil {
		return userID, nil
	}

	err = m.DB.QueryRow(`
        INSERT INTO users 
            (username, email, provider, provider_id, role, created_at, password_hash) 
        VALUES ($1, $2, $3, $4, 'user', CURRENT_TIMESTAMP, 'oauth')
        RETURNING user_id`,
		username, email, provider, provider_id).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (m *UserModel) PromoteUser(userID int) error {
	_, err := m.DB.Exec("UPDATE users SET role = 'moderator' WHERE user_id = $1", userID)
	return err
}

func (m *UserModel) DemoteUser(userID int) error {
	_, err := m.DB.Exec("UPDATE users SET role = 'user' WHERE user_id = $1", userID)
	return err
}
func (m *UserModel) ApplyForModerator(userID int) error {
	stmt := `UPDATE users SET role = 'pending_moderator' WHERE user_id = $1 AND role = 'user'`
	_, err := m.DB.Exec(stmt, userID)
	return err
}

func (m *UserModel) GetPendingModerators() ([]*User, error) {
	stmt := `SELECT user_id, username, email, role 
             FROM users 
             WHERE role IN ('pending_moderator', 'moderator')`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.username, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
