package domain

import "context"

type Picture struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	UserID       int64  `gorm:"type:bigint unsigned;not null;index"`
	Prompt       string `gorm:"type:text"` // 白底/透明图基础描述
	ScenePrompt  string `gorm:"type:text"` // 场景图用户描述
	EffectPrompt string `gorm:"type:text"` // 效果图用户描述
	WhiteBG      string `gorm:"type:text"`
	Transparent  string `gorm:"type:text"`
	Scene1       string `gorm:"type:text"`
	Scene2       string `gorm:"type:text"`
	EffectImage  string `gorm:"type:text"`
}

type PictureRepository interface {
	Create(ctx context.Context, p *Picture) error
	GetByUserID(ctx context.Context, userID int64) ([]*Picture, error)
}

type ImageGeneratorClient interface {
	Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
}

type GenerateImageRequest struct {
	ImageBase64    string
	Model          string
	Prompt         string // 白底/透明图基础描述
	ScenePrompt    string // 场景图用户描述（scene1/scene2 共用）
	EffectPrompt   string // 效果图用户描述
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Seed           *int
}

type FivePack struct {
	WhiteBG     string   `json:"white_bg"`
	Transparent string   `json:"transparent"`
	SceneImages []string `json:"scene_images"`
	EffectImage string   `json:"effect_image"`
}

type GenerateImageResult struct {
	ImageURL  string
	ImageURLs []string
	FivePack  *FivePack

	RequestID  string
	Width      int
	Height     int
	ImageCount int
}
