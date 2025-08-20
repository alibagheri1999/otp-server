package database

import (
	"context"
	"database/sql"
	"strings"

	"otp-server/internal/domain/entities"
	"otp-server/internal/domain/errors"
	"otp-server/internal/domain/repositories"

	"github.com/lib/pq"
)

// UserRepository implements the UserRepository interface using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(pool *PostgresPool) repositories.UserRepository {
	return &UserRepository{
		db: pool.db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (phone_number, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		user.PhoneNumber,
		user.Name,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return errors.NewAlreadyExists("user").WithError(err)
			case "23514":
				return errors.NewConstraintViolation("user", pqErr.Detail).WithError(err)
			default:
				return errors.NewDatabaseError("create user", err)
			}
		}
		return errors.NewDatabaseError("create user", err)
	}

	user.ID = id
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*entities.User, error) {
	query := `
		SELECT id, phone_number, name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("user")
		}
		return nil, errors.NewDatabaseError("get user by ID", err)
	}

	return &user, nil
}

// GetByPhoneNumber retrieves a user by phone number
func (r *UserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error) {
	query := `
		SELECT id, phone_number, name, role, is_active, created_at, updated_at
		FROM users WHERE phone_number = $1
	`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, phoneNumber).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("user")
		}
		return nil, errors.NewDatabaseError("get user by phone number", err)
	}

	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET name = $1, role = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Name,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("update user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFound("user")
	}

	return nil
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("delete user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFound("user")
	}

	return nil
}

// GetUsers retrieves a paginated list of users
func (r *UserRepository) GetUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	query := `
		SELECT id, phone_number, name, role, is_active, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError("get users", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.PhoneNumber,
			&user.Name,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("scan user", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("iterate users", err)
	}

	return users, nil
}

// GetTotalCount retrieves the total number of users
func (r *UserRepository) GetTotalCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseError("get user count", err)
	}

	return count, nil
}

// SearchUsers searches users by phone number or name
func (r *UserRepository) SearchUsers(ctx context.Context, query string) ([]*entities.User, error) {
	searchQuery := `
		SELECT id, phone_number, name, role, is_active, created_at, updated_at
		FROM users 
		WHERE phone_number ILIKE $1 OR name ILIKE $1
		ORDER BY created_at DESC
		LIMIT 50
	`

	searchPattern := "%" + strings.ToLower(query) + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern)
	if err != nil {
		return nil, errors.NewDatabaseError("search users", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.PhoneNumber,
			&user.Name,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("scan user", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("iterate search results", err)
	}

	return users, nil
}

// GetUsersWithQuery retrieves users with optional search and pagination in one query
func (r *UserRepository) GetUsersWithQuery(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error) {
	var baseQuery string
	var countQuery string
	var args []interface{}

	if query != "" {
		baseQuery = `
			SELECT id, phone_number, name, role, is_active, created_at, updated_at
			FROM users 
			WHERE phone_number ILIKE $1 OR name ILIKE $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		countQuery = `
			SELECT COUNT(*)
			FROM users 
			WHERE phone_number ILIKE $1 OR name ILIKE $1
		`
		searchPattern := "%" + strings.ToLower(query) + "%"
		args = []interface{}{searchPattern, limit, offset}
	} else {
		baseQuery = `
			SELECT id, phone_number, name, role, is_active, created_at, updated_at
			FROM users 
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		countQuery = `SELECT COUNT(*) FROM users`
		args = []interface{}{limit, offset}
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("get user count", err)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("get users with query", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID,
			&user.PhoneNumber,
			&user.Name,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewDatabaseError("scan user", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, errors.NewDatabaseError("iterate users", err)
	}

	return users, total, nil
}
