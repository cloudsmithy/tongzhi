package repository

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wechat-notification/models"

	_ "github.com/mattn/go-sqlite3"
)

// Repository errors
var (
	ErrNotFound        = errors.New("recipient not found")
	ErrDuplicateOpenID = errors.New("openid already exists")
)

// SQLiteRepository handles database operations
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

// initTables creates the necessary tables
func (r *SQLiteRepository) initTables() error {
	recipientsQuery := `
	CREATE TABLE IF NOT EXISTS recipients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		open_id TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := r.db.Exec(recipientsQuery); err != nil {
		return err
	}

	configQuery := `
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`
	if _, err := r.db.Exec(configQuery); err != nil {
		return err
	}

	templatesQuery := `
	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		template_id TEXT NOT NULL,
		name TEXT NOT NULL
	)`
	_, err := r.db.Exec(templatesQuery)
	return err
}

// Close closes the database connection
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// Create adds a new recipient to the database
func (r *SQLiteRepository) Create(recipient *models.Recipient) error {
	// Check for duplicate OpenID
	exists, err := r.OpenIDExists(recipient.OpenID)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateOpenID
	}

	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO recipients (open_id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
		recipient.OpenID, recipient.Name, now, now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	recipient.ID = id
	recipient.CreatedAt = now
	recipient.UpdatedAt = now
	return nil
}

// GetAll retrieves all recipients from the database
func (r *SQLiteRepository) GetAll() ([]models.Recipient, error) {
	rows, err := r.db.Query("SELECT id, open_id, name, created_at, updated_at FROM recipients ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []models.Recipient
	for rows.Next() {
		var rec models.Recipient
		if err := rows.Scan(&rec.ID, &rec.OpenID, &rec.Name, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		recipients = append(recipients, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil
	if recipients == nil {
		recipients = []models.Recipient{}
	}

	return recipients, nil
}

// GetByID retrieves a recipient by ID
func (r *SQLiteRepository) GetByID(id int64) (*models.Recipient, error) {
	var rec models.Recipient
	err := r.db.QueryRow(
		"SELECT id, open_id, name, created_at, updated_at FROM recipients WHERE id = ?",
		id,
	).Scan(&rec.ID, &rec.OpenID, &rec.Name, &rec.CreatedAt, &rec.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

// Update updates an existing recipient
func (r *SQLiteRepository) Update(recipient *models.Recipient) error {
	// Check if recipient exists
	_, err := r.GetByID(recipient.ID)
	if err != nil {
		return err
	}

	// Check if new OpenID conflicts with another recipient
	var existingID int64
	err = r.db.QueryRow("SELECT id FROM recipients WHERE open_id = ? AND id != ?", recipient.OpenID, recipient.ID).Scan(&existingID)
	if err == nil {
		return ErrDuplicateOpenID
	}
	if err != sql.ErrNoRows {
		return err
	}

	now := time.Now()
	_, err = r.db.Exec(
		"UPDATE recipients SET open_id = ?, name = ?, updated_at = ? WHERE id = ?",
		recipient.OpenID, recipient.Name, now, recipient.ID,
	)
	if err != nil {
		return err
	}

	recipient.UpdatedAt = now
	return nil
}

// Delete removes a recipient by ID
func (r *SQLiteRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM recipients WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// OpenIDExists checks if an OpenID already exists in the database
func (r *SQLiteRepository) OpenIDExists(openID string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM recipients WHERE open_id = ?", openID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}


// GetWeChatConfig retrieves WeChat configuration from database
func (r *SQLiteRepository) GetWeChatConfig() (*models.WeChatConfig, error) {
	config := &models.WeChatConfig{}
	
	rows, err := r.db.Query("SELECT key, value FROM config WHERE key IN ('wechat_app_id', 'wechat_app_secret', 'wechat_template_id')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		switch key {
		case "wechat_app_id":
			config.AppID = value
		case "wechat_app_secret":
			config.AppSecret = value
		case "wechat_template_id":
			config.TemplateID = value
		}
	}

	return config, rows.Err()
}

// SaveWeChatConfig saves WeChat configuration to database
func (r *SQLiteRepository) SaveWeChatConfig(config *models.WeChatConfig) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec("wechat_app_id", config.AppID); err != nil {
		return err
	}
	if _, err := stmt.Exec("wechat_app_secret", config.AppSecret); err != nil {
		return err
	}
	if _, err := stmt.Exec("wechat_template_id", config.TemplateID); err != nil {
		return err
	}

	return tx.Commit()
}


// GetConfig retrieves a config value by key
func (r *SQLiteRepository) GetConfig(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig saves a config value
func (r *SQLiteRepository) SetConfig(key, value string) error {
	_, err := r.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}

// GetByIDs retrieves recipients by their IDs
func (r *SQLiteRepository) GetByIDs(ids []int64) ([]models.Recipient, error) {
	if len(ids) == 0 {
		return []models.Recipient{}, nil
	}

	// Build placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "SELECT id, open_id, name, created_at, updated_at FROM recipients WHERE id IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []models.Recipient
	for rows.Next() {
		var rec models.Recipient
		if err := rows.Scan(&rec.ID, &rec.OpenID, &rec.Name, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		recipients = append(recipients, rec)
	}

	if recipients == nil {
		recipients = []models.Recipient{}
	}
	return recipients, rows.Err()
}


// CreateTemplate creates a new message template
func (r *SQLiteRepository) CreateTemplate(template *models.MessageTemplate) error {
	result, err := r.db.Exec(
		"INSERT INTO templates (key, template_id, name) VALUES (?, ?, ?)",
		template.Key, template.TemplateID, template.Name,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	template.ID = id
	return nil
}

// GetAllTemplates retrieves all templates
func (r *SQLiteRepository) GetAllTemplates() ([]models.MessageTemplate, error) {
	rows, err := r.db.Query("SELECT id, key, template_id, name FROM templates ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		if err := rows.Scan(&t.ID, &t.Key, &t.TemplateID, &t.Name); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if templates == nil {
		templates = []models.MessageTemplate{}
	}
	return templates, rows.Err()
}

// GetTemplateByKey retrieves a template by key
func (r *SQLiteRepository) GetTemplateByKey(key string) (*models.MessageTemplate, error) {
	var t models.MessageTemplate
	err := r.db.QueryRow("SELECT id, key, template_id, name FROM templates WHERE key = ?", key).
		Scan(&t.ID, &t.Key, &t.TemplateID, &t.Name)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &t, err
}

// DeleteTemplate deletes a template by ID
func (r *SQLiteRepository) DeleteTemplate(id int64) error {
	result, err := r.db.Exec("DELETE FROM templates WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
