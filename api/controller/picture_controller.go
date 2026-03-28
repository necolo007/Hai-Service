package controller

import (
	"Hai-Service/config"
	"Hai-Service/core/libx"
	"Hai-Service/core/store/mysql"
	"Hai-Service/repository"
	"Hai-Service/usecase"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PictureController struct {
	uc *usecase.PictureUsecase
}

func NewPictureController() *PictureController {
	db, _ := mysql.InitMySQL()

	repo := repository.NewPictureRepo(db)
	generator := usecase.NewDashScopeI2IClient(
		config.GetConfig().Picture.Endpoint,
		config.GetConfig().Picture.APIKey,
	)
	uc := usecase.NewPictureUsecase(repo, generator)
	return &PictureController{uc: uc}
}

func (p *PictureController) Register(r *gin.RouterGroup) {
	r.POST("/pictures/generate", p.Generate)
	r.GET("/pictures/me", p.GetByUserID)
}

type generatePictureForm struct {
	Prompt         string `form:"prompt"`        // 白底/透明图基础描述（可选，留空后端使用默认描述）
	ScenePrompt    string `form:"scene_prompt"`  // 场景图描述（scene1/scene2 共用，可选）
	EffectPrompt   string `form:"effect_prompt"` // 效果图描述（可选）
	NegativePrompt string `form:"negative_prompt"`
	Size           string `form:"size"`
	PromptExtend   *bool  `form:"prompt_extend"`
	Model          string `form:"model"`
	Seed           *int   `form:"seed"`
}

const dashScopeMinDim = 384

// resizeNearestNeighbor 用最近邻算法将图片缩放到指定尺寸。
func resizeNearestNeighbor(src image.Image, newW, newH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	srcB := src.Bounds()
	srcW, srcH := srcB.Dx(), srcB.Dy()
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			sx := x * srcW / newW
			sy := y * srcH / newH
			dst.Set(x, y, src.At(srcB.Min.X+sx, srcB.Min.Y+sy))
		}
	}
	return dst
}

// ensureMinDimension 若图片任意边小于 dashScopeMinDim，则等比放大至满足要求。
// 返回处理后的字节和 MIME 类型（统一输出 image/jpeg）。
func ensureMinDimension(b []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		// 无法解码则原样返回（交由 API 报错）
		return b, nil
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w >= dashScopeMinDim && h >= dashScopeMinDim {
		return b, nil
	}

	// 找到最短边，计算等比放大倍数使两边都 >= minDim
	scale := math.Ceil(float64(dashScopeMinDim)/float64(min(w, h))*100) / 100
	newW := int(math.Ceil(float64(w) * scale))
	newH := int(math.Ceil(float64(h) * scale))

	resized := resizeNearestNeighbor(img, newW, newH)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 95}); err != nil {
		return nil, fmt.Errorf("re-encode image: %w", err)
	}
	return buf.Bytes(), nil
}

func fileToDataBase64(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	b, err = ensureMinDimension(b)
	if err != nil {
		return "", err
	}

	// 统一声明为 image/jpeg（重编后均为 JPEG）
	enc := base64.StdEncoding.EncodeToString(b)
	return fmt.Sprintf("data:image/jpeg;base64,%s", enc), nil
}

func (p *PictureController) Generate(c *gin.Context) {
	userID := libx.Uid(c)

	var form generatePictureForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fh, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	imgBase64, err := fileToDataBase64(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promptExtend := true
	if form.PromptExtend != nil {
		promptExtend = *form.PromptExtend
	}

	_, genRes, err := p.uc.GenerateAndSave(c.Request.Context(), usecase.GeneratePictureInput{
		UserID:         int64(userID),
		ImageBase64:    imgBase64,
		Prompt:         form.Prompt,
		ScenePrompt:    form.ScenePrompt,
		EffectPrompt:   form.EffectPrompt,
		NegativePrompt: form.NegativePrompt,
		Size:           form.Size,
		PromptExtend:   promptExtend,
		Watermark:      false,
		Model:          form.Model,
		Seed:           form.Seed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	libx.Ok(c, "创建成功", gin.H{
		"request_id": genRes.RequestID,
		"images":     genRes.ImageURLs,
		"five_pack":  genRes.FivePack,
		"usage": gin.H{
			"width":       genRes.Width,
			"height":      genRes.Height,
			"image_count": genRes.ImageCount,
		},
	})
}

func (p *PictureController) GetByUserID(c *gin.Context) {
	userID := libx.Uid(c)

	pics, err := p.uc.GetByUserID(c.Request.Context(), int64(userID))
	if err != nil {
		libx.Err(c, http.StatusInternalServerError, "查询失败", err)
		return
	}

	type picItem struct {
		ID           int64    `json:"id"`
		Prompt       string   `json:"prompt"`
		ScenePrompt  string   `json:"scene_prompt"`
		EffectPrompt string   `json:"effect_prompt"`
		WhiteBG      string   `json:"white_bg"`
		Transparent  string   `json:"transparent"`
		SceneImages  []string `json:"scene_images"`
		EffectImage  string   `json:"effect_image"`
	}
	items := make([]picItem, 0, len(pics))
	for _, pic := range pics {
		items = append(items, picItem{
			ID:           pic.ID,
			Prompt:       pic.Prompt,
			ScenePrompt:  pic.ScenePrompt,
			EffectPrompt: pic.EffectPrompt,
			WhiteBG:      pic.WhiteBG,
			Transparent:  pic.Transparent,
			SceneImages:  []string{pic.Scene1, pic.Scene2},
			EffectImage:  pic.EffectImage,
		})
	}

	libx.Ok(c, "查询成功", gin.H{
		"total":    len(items),
		"pictures": items,
	})
}
