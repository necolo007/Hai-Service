package repository

import (
	"Hai-Service/domain"
	"context"
	"gorm.io/gorm"
)

type PictureRepo struct {
	db *gorm.DB
}

func NewPictureRepo(db *gorm.DB) domain.PictureRepository {
	return &PictureRepo{db: db}
}

func (r *PictureRepo) Create(ctx context.Context, p *domain.Picture) error {
	// 使用 WithContext 保持 context 传播
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PictureRepo) GetByUserID(ctx context.Context, userID int64) ([]*domain.Picture, error) {
	var pics []*domain.Picture
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&pics).Error; err != nil {
		return nil, err
	}
	return pics, nil
}
