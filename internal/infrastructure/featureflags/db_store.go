package featureflags

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const cacheTTL = 5 * time.Minute

// DBStore is the authoritative PostgreSQL-backed feature flag store with
// Redis caching for fast reads.
type DBStore struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewDBStore(db *gorm.DB, cache *redis.Client) Store {
	return &DBStore{db: db, cache: cache}
}

func (s *DBStore) cacheKey(tenantID *uint, name string) string {
	if tenantID == nil {
		return fmt.Sprintf("ff:global:%s", name)
	}
	return fmt.Sprintf("ff:%d:%s", *tenantID, name)
}

func (s *DBStore) IsEnabled(ctx context.Context, tenantID *uint, name string) (bool, error) {
	flag, err := s.Get(ctx, tenantID, name)
	if err != nil || flag == nil {
		return false, err
	}
	return flag.Enabled, nil
}

func (s *DBStore) IsEnabledForUser(ctx context.Context, tenantID *uint, userID uint, name string) (bool, error) {
	flag, err := s.Get(ctx, tenantID, name)
	if err != nil || flag == nil {
		return false, err
	}
	if !flag.Enabled {
		return false, nil
	}

	if len(flag.Rules) == 0 || string(flag.Rules) == "{}" {
		return flag.Enabled, nil
	}

	var rules FlagRules
	if err := json.Unmarshal(flag.Rules, &rules); err != nil {
		return flag.Enabled, nil
	}

	// User-level override
	for _, uid := range rules.UserIDs {
		if uid == userID {
			return true, nil
		}
	}

	return flag.Enabled, nil
}

func (s *DBStore) Get(ctx context.Context, tenantID *uint, name string) (*FeatureFlag, error) {
	key := s.cacheKey(tenantID, name)

	// Try cache first
	if data, err := s.cache.Get(ctx, key).Bytes(); err == nil {
		var flag FeatureFlag
		if json.Unmarshal(data, &flag) == nil {
			return &flag, nil
		}
	}

	var flag FeatureFlag
	q := s.db.WithContext(ctx).Where("name = ?", name)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	} else {
		q = q.Where("tenant_id IS NULL")
	}

	err := q.First(&flag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(&flag); err == nil {
		_ = s.cache.Set(ctx, key, data, cacheTTL).Err()
	}

	return &flag, nil
}

func (s *DBStore) List(ctx context.Context, tenantID *uint) ([]*FeatureFlag, error) {
	var flags []*FeatureFlag
	q := s.db.WithContext(ctx)
	if tenantID != nil {
		q = q.Where("tenant_id = ? OR tenant_id IS NULL", *tenantID)
	} else {
		q = q.Where("tenant_id IS NULL")
	}
	err := q.Order("name ASC").Find(&flags).Error
	return flags, err
}

func (s *DBStore) Set(ctx context.Context, tenantID *uint, name string, enabled bool) error {
	q := s.db.WithContext(ctx).Model(&FeatureFlag{}).Where("name = ?", name)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	} else {
		q = q.Where("tenant_id IS NULL")
	}
	if err := q.Update("enabled", enabled).Error; err != nil {
		return err
	}
	return s.Invalidate(ctx, tenantID, name)
}

func (s *DBStore) Create(ctx context.Context, flag *FeatureFlag) error {
	return s.db.WithContext(ctx).Create(flag).Error
}

func (s *DBStore) Delete(ctx context.Context, tenantID *uint, name string) error {
	q := s.db.WithContext(ctx).Where("name = ?", name)
	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	} else {
		q = q.Where("tenant_id IS NULL")
	}
	if err := q.Delete(&FeatureFlag{}).Error; err != nil {
		return err
	}
	return s.Invalidate(ctx, tenantID, name)
}

func (s *DBStore) Invalidate(ctx context.Context, tenantID *uint, name string) error {
	return s.cache.Del(ctx, s.cacheKey(tenantID, name)).Err()
}
