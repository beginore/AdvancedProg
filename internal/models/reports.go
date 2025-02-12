package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Import PostgreSQL driver
)

type Report struct {
	ID         int
	PostID     int
	ReporterID int
	Reason     string
	CreatedAt  time.Time
	Answer     string
	AdminID    int
	Solved     int
}

type ReportModel struct {
	DB *sql.DB
}

func (m *ReportModel) Get(id int) (*Report, error) {
	stmt := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer, solved FROM reports WHERE id = $1`

	row := m.DB.QueryRow(stmt, id)

	r := &Report{}
	err := row.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer, &r.Solved)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return r, nil
}

func (m *ReportModel) Create(postID, reporterID int, reason string) error {
	query := `INSERT INTO reports (post_id, reporter_id, reason, created_at) VALUES ($1, $2, $3, NOW())`
	_, err := m.DB.Exec(query, postID, reporterID, reason)
	if err != nil {
		return err
	}
	return nil
}

func (m *ReportModel) Answer(reportID, adminID int, answer string) error {
	var exists bool
	err := m.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM reports WHERE id = $1)", reportID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ошибка: отчёт с ID %d не найден", reportID)
	}

	query := `UPDATE reports SET admin_id = $1, answer = $2, solved = 1 WHERE id = $3`
	_, err = m.DB.Exec(query, adminID, answer, reportID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ReportModel) GetUnsolved() ([]*Report, error) {
	query := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer, solved FROM reports WHERE solved = 0`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err = rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer, &r.Solved)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reports, nil
}

func (m *ReportModel) GetSolved() ([]*Report, error) {
	query := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer, solved FROM reports WHERE solved = 1`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err := rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer, &r.Solved)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}

func (m *ReportModel) GetAll() ([]*Report, error) {
	query := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer, solved FROM reports`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err := rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer, &r.Solved)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}
