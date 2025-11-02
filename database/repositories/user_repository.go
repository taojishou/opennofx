package repositories

import (
	"database/sql"
	"nofx/database/models"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *models.User) (int64, error) {
	query := `
		INSERT INTO users (username, email, password, role, is_active)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, user.Username, user.Email, user.Password, user.Role, user.IsActive)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at, is_active
		FROM users WHERE id = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at, is_active
		FROM users WHERE username = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at, is_active
		FROM users WHERE email = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET email = ?, password = ?, role = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, user.Email, user.Password, user.Role, user.IsActive, user.ID)
	return err
}

// GetAll 获取所有用户
func (r *UserRepository) GetAll() ([]*models.User, error) {
	query := `
		SELECT id, username, email, password, role, created_at, updated_at, is_active
		FROM users ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Role, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// CreateSession 创建会话
func (r *UserRepository) CreateSession(session *models.Session) (int64, error) {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`
	result, err := r.db.Exec(query, session.UserID, session.Token, session.ExpiresAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetSessionByToken 根据Token获取会话
func (r *UserRepository) GetSessionByToken(token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
	`
	session := &models.Session{}
	err := r.db.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token,
		&session.ExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// DeleteSession 删除会话
func (r *UserRepository) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := r.db.Exec(query, token)
	return err
}
