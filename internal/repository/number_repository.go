package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type NumberRepository struct {
	db *gorm.DB
}

func NewNumberRepository(db *gorm.DB) *NumberRepository {
	return &NumberRepository{db: db}
}

type Plate struct {
	ID         int64     `gorm:"primaryKey"`
	Number     string    `gorm:"not null"`
	Normalized string    `gorm:"not null;uniqueIndex"`
	Country    *string
	Region     *string
	CreatedAt  time.Time
}

type List struct {
	ID          int64     `gorm:"primaryKey"`
	Name        string    `gorm:"not null;uniqueIndex"`
	Type        string    `gorm:"not null"`
	Description *string
	CreatedAt   time.Time
}

type ListItem struct {
	ListID    int64     `gorm:"primaryKey"`
	PlateID   int64     `gorm:"primaryKey"`
	Note      *string
	CreatedAt time.Time
}

func (r *NumberRepository) GetOrCreatePlate(ctx context.Context, normalized, original string) (int64, error) {
	var plate Plate
	err := r.db.WithContext(ctx).Where("normalized = ?", normalized).First(&plate).Error
	if err == nil {
		return plate.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	plate = Plate{
		Number:     original,
		Normalized: normalized,
		CreatedAt:  time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(&plate).Error; err != nil {
		return 0, err
	}
	return plate.ID, nil
}

func (r *NumberRepository) FindPlatesByNormalized(ctx context.Context, normalized string) ([]Plate, error) {
	var plates []Plate
	err := r.db.WithContext(ctx).
		Where("normalized = ?", normalized).
		Find(&plates).Error
	return plates, err
}

func (r *NumberRepository) FindListsForPlate(ctx context.Context, plateID int64) ([]ListHit, error) {
	var hits []ListHit

	err := r.db.WithContext(ctx).
		Table("list_items").
		Select("lists.id as list_id, lists.name as list_name, lists.type as list_type").
		Joins("JOIN lists ON list_items.list_id = lists.id").
		Where("list_items.plate_id = ?", plateID).
		Scan(&hits).Error

	if err != nil {
		return nil, err
	}

	return hits, nil
}

func (r *NumberRepository) AddPlateToList(ctx context.Context, listID, plateID int64, note *string) error {
	item := ListItem{
		ListID:    listID,
		PlateID:   plateID,
		Note:      note,
		CreatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(&item).Error
}

func (r *NumberRepository) RemovePlateFromList(ctx context.Context, listID, plateID int64) error {
	return r.db.WithContext(ctx).
		Where("list_id = ? AND plate_id = ?", listID, plateID).
		Delete(&ListItem{}).Error
}

func (r *NumberRepository) GetList(ctx context.Context, listID int64) (*List, error) {
	var list List
	err := r.db.WithContext(ctx).Where("id = ?", listID).First(&list).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *NumberRepository) GetListByName(ctx context.Context, name string) (*List, error) {
	var list List
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&list).Error
	if err != nil {
		return nil, err
	}
	return &list, nil
}

type ListHit struct {
	ListID   int64  `json:"list_id"`
	ListName string `json:"list_name"`
	ListType string `json:"list_type"`
}

