package usecase

import (
	"Hai-Service/domain"
	"context"
	"errors"
)

type PictureUsecase struct {
	repo      domain.PictureRepository
	generator domain.ImageGeneratorClient
}

func NewPictureUsecase(repo domain.PictureRepository, generator domain.ImageGeneratorClient) *PictureUsecase {
	return &PictureUsecase{repo: repo, generator: generator}
}

type GeneratePictureInput struct {
	UserID         int64
	ImageBase64    string
	Prompt         string // 白底/透明图基础描述
	ScenePrompt    string // 场景图用户描述
	EffectPrompt   string // 效果图用户描述
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Model          string
	Seed           *int
}

func (u *PictureUsecase) GenerateAndSave(ctx context.Context, in GeneratePictureInput) (*domain.Picture, *domain.GenerateImageResult, error) {
	if in.ImageBase64 == "" {
		return nil, nil, errors.New("image base64 empty")
	}

	model := in.Model
	if model == "" {
		model = "wan2.5-i2i-preview"
	}
	size := in.Size
	if size == "" {
		size = "1280*1280"
	}

	res, err := u.generator.Generate(ctx, domain.GenerateImageRequest{
		ImageBase64:    in.ImageBase64,
		Model:          model,
		Prompt:         in.Prompt,
		ScenePrompt:    in.ScenePrompt,
		EffectPrompt:   in.EffectPrompt,
		NegativePrompt: in.NegativePrompt,
		Size:           size,
		PromptExtend:   in.PromptExtend,
		Watermark:      in.Watermark,
		Seed:           in.Seed,
	})
	if err != nil {
		return nil, nil, err
	}

	fp := res.FivePack
	var whiteBG, transparent, scene1, scene2, effectImage string
	if fp != nil {
		whiteBG = fp.WhiteBG
		transparent = fp.Transparent
		if len(fp.SceneImages) > 0 {
			scene1 = fp.SceneImages[0]
		}
		if len(fp.SceneImages) > 1 {
			scene2 = fp.SceneImages[1]
		}
		effectImage = fp.EffectImage
	}

	p := &domain.Picture{
		UserID:       in.UserID,
		Prompt:       in.Prompt,
		ScenePrompt:  in.ScenePrompt,
		EffectPrompt: in.EffectPrompt,
		WhiteBG:      whiteBG,
		Transparent:  transparent,
		Scene1:       scene1,
		Scene2:       scene2,
		EffectImage:  effectImage,
	}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, nil, err
	}
	return p, res, nil
}

func (u *PictureUsecase) GetByUserID(ctx context.Context, userID int64) ([]*domain.Picture, error) {
	return u.repo.GetByUserID(ctx, userID)
}
