package models

import (
	"database/sql"
	"strings"
)

// Category represents a category in the database
type Category struct {
	ID   int
	Name string
}

// CategoryModel provides methods for interacting with the categories table
type CategoryModel struct {
	DB *sql.DB
}

// GetAll retrieves all categories from the database
func (m *CategoryModel) GetAll() ([]*Category, error) {
	stmt := `
        SELECT id, name 
        FROM categories
        ORDER BY name ASC
    `
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		c := &Category{}
		err = rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// Insert adds a new category to the database
func (m *CategoryModel) Insert(name string) error {
	stmt := `
        INSERT INTO categories (name) 
        VALUES ($1)
        ON CONFLICT (name) DO NOTHING
    `
	_, err := m.DB.Exec(stmt, name)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrDuplicateCategory
		}
		return err
	}
	return nil
}

// Update modifies an existing category
func (m *CategoryModel) Update(id int, newName string) error {
	stmt := `
        UPDATE categories 
        SET name = $1 
        WHERE id = $2
    `
	_, err := m.DB.Exec(stmt, newName, id)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrDuplicateCategory
		}
		return err
	}
	return nil
}

// Delete removes a category by ID
func (m *CategoryModel) Delete(id int) error {
	stmt := `
        DELETE FROM categories 
        WHERE id = $1
    `
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}
