package contacts

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	CountParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool) (int64, error)
	ListParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool, limit, offset int) ([]Party, error)
	GetParty(ctx context.Context, tenantID, id uint) (*Party, error)
	CreateParty(ctx context.Context, p *Party) error
	UpdateParty(ctx context.Context, p *Party) error
	DeleteParty(ctx context.Context, tenantID, id uint) error

	ListContactPersons(ctx context.Context, tenantID, partyID uint) ([]ContactPerson, error)
	GetContactPerson(ctx context.Context, tenantID, id uint) (*ContactPerson, error)
	CreateContactPerson(ctx context.Context, cp *ContactPerson) error
	UpdateContactPerson(ctx context.Context, cp *ContactPerson) error
	DeleteContactPerson(ctx context.Context, tenantID, id uint) error

	ListAddresses(ctx context.Context, tenantID, partyID uint) ([]PartyAddress, error)
	GetAddress(ctx context.Context, tenantID, id uint) (*PartyAddress, error)
	CreateAddress(ctx context.Context, a *PartyAddress) error
	UpdateAddress(ctx context.Context, a *PartyAddress) error
	DeleteAddress(ctx context.Context, tenantID, id uint) error
}

type dbRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &dbRepository{db: db} }

func (r *dbRepository) CountParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool) (int64, error) {
	q := r.db.WithContext(ctx).Model(&Party{}).Where("tenant_id = ?", tenantID)
	if partyType != "" {
		q = q.Where("party_type = ? OR party_type = 'BOTH'", partyType)
	}
	if activeOnly {
		q = q.Where("is_active = true")
	}
	var count int64
	return count, q.Count(&count).Error
}

func (r *dbRepository) ListParties(ctx context.Context, tenantID uint, partyType string, activeOnly bool, limit, offset int) ([]Party, error) {
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if partyType != "" {
		q = q.Where("party_type = ? OR party_type = 'BOTH'", partyType)
	}
	if activeOnly {
		q = q.Where("is_active = true")
	}
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	var rows []Party
	return rows, q.Order("name").Find(&rows).Error
}

func (r *dbRepository) GetParty(ctx context.Context, tenantID, id uint) (*Party, error) {
	var p Party
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *dbRepository) CreateParty(ctx context.Context, p *Party) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *dbRepository) UpdateParty(ctx context.Context, p *Party) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *dbRepository) DeleteParty(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&Party{}).Error
}

func (r *dbRepository) ListContactPersons(ctx context.Context, tenantID, partyID uint) ([]ContactPerson, error) {
	var rows []ContactPerson
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND party_id = ?", tenantID, partyID).
		Order("is_primary DESC, name").
		Find(&rows).Error
}

func (r *dbRepository) GetContactPerson(ctx context.Context, tenantID, id uint) (*ContactPerson, error) {
	var cp ContactPerson
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&cp).Error
	if err != nil {
		return nil, err
	}
	return &cp, nil
}

func (r *dbRepository) CreateContactPerson(ctx context.Context, cp *ContactPerson) error {
	return r.db.WithContext(ctx).Create(cp).Error
}

func (r *dbRepository) UpdateContactPerson(ctx context.Context, cp *ContactPerson) error {
	return r.db.WithContext(ctx).Save(cp).Error
}

func (r *dbRepository) DeleteContactPerson(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&ContactPerson{}).Error
}

func (r *dbRepository) ListAddresses(ctx context.Context, tenantID, partyID uint) ([]PartyAddress, error) {
	var rows []PartyAddress
	return rows, r.db.WithContext(ctx).
		Where("tenant_id = ? AND party_id = ?", tenantID, partyID).
		Order("is_primary DESC").
		Find(&rows).Error
}

func (r *dbRepository) GetAddress(ctx context.Context, tenantID, id uint) (*PartyAddress, error) {
	var a PartyAddress
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *dbRepository) CreateAddress(ctx context.Context, a *PartyAddress) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *dbRepository) UpdateAddress(ctx context.Context, a *PartyAddress) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *dbRepository) DeleteAddress(ctx context.Context, tenantID, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&PartyAddress{}).Error
}
