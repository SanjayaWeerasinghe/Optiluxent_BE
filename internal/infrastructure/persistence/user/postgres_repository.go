package user

import (
	"context"
	stderrors "errors"

	domain "erp-system/internal/domain/user"

	"gorm.io/gorm"
)

// PostgresRepository implements domain/user.Repository using GORM.
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository returns a new PostgresRepository.
func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&u).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&u).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": gorm.Expr("NOW()"),
			"status":     domain.StatusInactive,
		}).Error
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.User{}).Where("deleted_at IS NULL")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *PostgresRepository) ChangePassword(ctx context.Context, id uint, newPasswordHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("password_hash", newPasswordHash).Error
}

// SetRole replaces all role assignments for a user with a single role inside a
// transaction. Both the user_roles join (used by Casbin's DB adapter) and the
// denormalised users.role string are updated together — they must agree.
func (r *PostgresRepository) SetRole(ctx context.Context, userID, roleID uint, roleName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM user_roles WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`,
			userID, roleID,
		).Error; err != nil {
			return err
		}
		return tx.Model(&domain.User{}).
			Where("id = ?", userID).
			Update("role", roleName).Error
	})
}
