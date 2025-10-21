package repositories

import (
	"database/sql"
	"fmt"
	"gosprints/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByUsername(username string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	
	err := r.db.QueryRow(
		`INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`,
		user.Username, user.Password,
	).Scan(&user.ID)
	
	if err != nil {
		fmt.Printf("User Create ERROR: %v\n", err)
		return err
	}
	
	fmt.Printf("User created: ID=%d, Username=%s\n", user.ID, user.Username)
	return nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	
	var user models.User
	err := r.db.QueryRow(
		`SELECT id, username, password FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password)
	
	if err != nil {
		fmt.Printf("User GetByUsername ERROR: %v\n", err)
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	fmt.Printf("User found: ID=%d, Username=%s\n", user.ID, user.Username)
	return &user, nil
}