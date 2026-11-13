package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"victor-tool-collection/backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ScoreHandler manages the score library: scores (metadata) plus their images.
// Images are either mirrored local files (served by nginx at /scores/files/)
// or remote URLs kept as-is.
type ScoreHandler struct {
	DB  *gorm.DB
	Dir string // absolute path of the local storage directory
}

const (
	maxImageSize   = 20 << 20 // 20 MiB per image
	maxImagesScore = 60       // hard cap per score
)

var allowedImageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true,
}

var extByMime = map[string]string{
	"image/png": ".png", "image/jpeg": ".jpg", "image/webp": ".webp", "image/gif": ".gif",
}

/* ------------------------------------------------------------------ */
/* payload shaping                                                     */
/* ------------------------------------------------------------------ */

func scoreJSON(s models.Score) gin.H {
	imgs := make([]gin.H, 0, len(s.Images))
	for _, im := range s.Images {
		imgs = append(imgs, gin.H{
			"id":         im.ID,
			"url":        im.URL,
			"source":     im.Source,
			"source_url": im.SourceURL,
			"sort_order": im.SortOrder,
		})
	}
	return gin.H{
		"id":         s.ID,
		"title":      s.Title,
		"composer":   s.Composer,
		"type":       s.Type,
		"note":       s.Note,
		"created_at": s.CreatedAt,
		"images":     imgs,
		"image_count": len(imgs),
		"cover":      firstImageURL(s.Images),
	}
}

func firstImageURL(imgs []models.ScoreImage) string {
	if len(imgs) == 0 {
		return ""
	}
	return imgs[0].URL
}

/* ------------------------------------------------------------------ */
/* endpoints                                                           */
/* ------------------------------------------------------------------ */

// List returns scores with their images, newest first.
// GET /api/scores?q=keyword&type=吉他谱
func (h *ScoreHandler) List(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	typ := strings.TrimSpace(c.Query("type"))

	tx := h.DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	})
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		tx = tx.Where("LOWER(title) LIKE ? OR LOWER(composer) LIKE ? OR LOWER(note) LIKE ?", like, like, like)
	}
	if typ != "" {
		tx = tx.Where("type = ?", typ)
	}

	var rows []models.Score
	if err := tx.Order("created_at DESC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "查询失败"})
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		items = append(items, scoreJSON(r))
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "items": items, "count": len(items), "query": q})
}

// Types lists the distinct score types in use (for filters).
// GET /api/scores/types
func (h *ScoreHandler) Types(c *gin.Context) {
	var types []string
	h.DB.Model(&models.Score{}).
		Where("type <> ''").
		Distinct().
		Pluck("type", &types)
	c.JSON(http.StatusOK, gin.H{"ok": true, "types": types})
}

type createScoreRequest struct {
	Title    string   `json:"title"`
	Composer string   `json:"composer"`
	Type     string   `json:"type"`
	Note     string   `json:"note"`
	Images   []string `json:"images"` // remote image URLs
}

// Create adds a score, optionally with remote images.
// POST /api/scores
func (h *ScoreHandler) Create(c *gin.Context) {
	var req createScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "乐谱名称不能为空"})
		return
	}
	if req.Type == "" {
		req.Type = "吉他谱"
	}
	if len(req.Images) > maxImagesScore {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "图片数量超出上限"})
		return
	}

	rec := models.Score{
		Title:    req.Title,
		Composer: strings.TrimSpace(req.Composer),
		Type:     strings.TrimSpace(req.Type),
		Note:     strings.TrimSpace(req.Note),
	}
	if err := h.DB.Create(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "写入数据库失败"})
		return
	}

	failures := make([]string, 0)
	for i, raw := range req.Images {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, err := h.addImageURL(rec.ID, raw, i); err != nil {
			failures = append(failures, raw+" → "+err.Error())
		}
	}

	h.DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).First(&rec, rec.ID)

	c.JSON(http.StatusOK, gin.H{"ok": true, "item": scoreJSON(rec), "failed": failures})
}

// AddImages appends images to an existing score.
// POST /api/scores/:id/images
// Accepts either multipart (field "files", repeatable) or JSON {"urls": [...]}.
func (h *ScoreHandler) AddImages(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var rec models.Score
	if err := h.DB.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "乐谱不存在"})
		return
	}

	start := h.nextSortOrder(rec.ID)
	added := 0
	failures := make([]string, 0)

	if strings.HasPrefix(c.ContentType(), "multipart/") {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "表单解析失败"})
			return
		}
		files := form.File["files"]
		if len(files) == 0 {
			files = form.File["file"]
		}
		for i, fh := range files {
			if err := h.saveMultipartImage(rec.ID, fh.Filename, fh.Size, func() (io.ReadCloser, error) {
				return fh.Open()
			}, start+i); err != nil {
				failures = append(failures, fh.Filename+" → "+err.Error())
				continue
			}
			added++
		}
	} else {
		var req struct {
			URLs []string `json:"urls"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
			return
		}
		for i, raw := range req.URLs {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			if _, err := h.addImageURL(rec.ID, raw, start+i); err != nil {
				failures = append(failures, raw+" → "+err.Error())
				continue
			}
			added++
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "added": added, "failed": failures})
}

// DeleteImage removes one image from a score (and its local file).
// DELETE /api/scores/images/:imageId
func (h *ScoreHandler) DeleteImage(c *gin.Context) {
	imgID, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}
	var im models.ScoreImage
	if err := h.DB.First(&im, imgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "图片不存在"})
		return
	}
	if err := h.DB.Delete(&im).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "删除失败"})
		return
	}
	h.removeLocalFile(im.StoredName)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Delete removes a score, all its image rows and any mirrored local files.
// DELETE /api/scores/:id
func (h *ScoreHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var rec models.Score
	if err := h.DB.Preload("Images").First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "乐谱不存在"})
		return
	}
	for _, im := range rec.Images {
		h.removeLocalFile(im.StoredName)
	}
	h.DB.Where("score_id = ?", rec.ID).Delete(&models.ScoreImage{})
	if err := h.DB.Delete(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UpdateScore edits title/composer/type/note.
// PATCH /api/scores/:id
func (h *ScoreHandler) UpdateScore(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var rec models.Score
	if err := h.DB.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "乐谱不存在"})
		return
	}
	var req struct {
		Title    *string `json:"title"`
		Composer *string `json:"composer"`
		Type     *string `json:"type"`
		Note     *string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "名称不能为空"})
			return
		}
		rec.Title = strings.TrimSpace(*req.Title)
	}
	if req.Composer != nil {
		rec.Composer = strings.TrimSpace(*req.Composer)
	}
	if req.Type != nil && strings.TrimSpace(*req.Type) != "" {
		rec.Type = strings.TrimSpace(*req.Type)
	}
	if req.Note != nil {
		rec.Note = strings.TrimSpace(*req.Note)
	}
	if err := h.DB.Save(&rec).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

/* ------------------------------------------------------------------ */
/* image plumbing                                                      */
/* ------------------------------------------------------------------ */

// addImageURL stores a remote URL. http:// URLs (and/or host-allowlist
// callers) are mirrored to local disk because an https page cannot render
// mixed-content images, and many score sites block hotlinking.
func (h *ScoreHandler) addImageURL(scoreID uint, raw string, order int) (models.ScoreImage, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return models.ScoreImage{}, fmt.Errorf("无效的图片链接")
	}

	im := models.ScoreImage{
		ScoreID:   scoreID,
		SourceURL: raw,
		SortOrder: order,
		Source:    "remote",
		URL:       raw,
	}

	// Mirror plain-http images so they survive on an https page.
	if u.Scheme == "http" {
		stored, err := h.mirrorRemote(raw)
		if err != nil {
			return models.ScoreImage{}, err
		}
		im.Source = "local"
		im.StoredName = stored
		im.URL = "/scores/files/" + stored
	}

	if err := h.DB.Create(&im).Error; err != nil {
		h.removeLocalFile(im.StoredName)
		return models.ScoreImage{}, fmt.Errorf("写入数据库失败")
	}
	return im, nil
}

// mirrorRemote downloads a remote image into the local store.
func (h *ScoreHandler) mirrorRemote(raw string) (string, error) {
	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return "", fmt.Errorf("下载失败")
	}
	// No Referer on purpose: score sites commonly allow empty referer only.
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,*/*")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("源站返回 %d", resp.StatusCode)
	}

	ext := extByMime[strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))]
	if ext == "" {
		ext = filepath.Ext(strings.Split(raw, "?")[0])
	}
	if !allowedImageExts[ext] {
		ext = ".jpg"
	}

	if err := os.MkdirAll(h.Dir, 0o755); err != nil {
		return "", fmt.Errorf("创建目录失败")
	}
	stored := randomScoreName(ext)
	dst := filepath.Join(h.Dir, stored)

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("保存失败")
	}
	defer f.Close()

	written, err := io.Copy(f, io.LimitReader(resp.Body, maxImageSize+1))
	if err != nil {
		os.Remove(dst)
		return "", fmt.Errorf("保存失败")
	}
	if written > maxImageSize {
		os.Remove(dst)
		return "", fmt.Errorf("图片超过 20MB")
	}
	return stored, nil
}

// saveMultipartImage writes one uploaded file to the local store and records it.
func (h *ScoreHandler) saveMultipartImage(scoreID uint, filename string, size int64, open func() (io.ReadCloser, error), order int) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageExts[ext] {
		return fmt.Errorf("不支持的格式（仅 PNG/JPG/WebP/GIF）")
	}
	if size > maxImageSize {
		return fmt.Errorf("图片超过 20MB")
	}
	if err := os.MkdirAll(h.Dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败")
	}
	stored := randomScoreName(ext)
	dst := filepath.Join(h.Dir, stored)

	src, err := open()
	if err != nil {
		return fmt.Errorf("读取失败")
	}
	defer src.Close()

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("保存失败")
	}
	if _, err := io.Copy(f, io.LimitReader(src, maxImageSize+1)); err != nil {
		f.Close()
		os.Remove(dst)
		return fmt.Errorf("保存失败")
	}
	f.Close()

	im := models.ScoreImage{
		ScoreID:    scoreID,
		URL:        "/scores/files/" + stored,
		Source:     "local",
		StoredName: stored,
		SortOrder:  order,
	}
	if err := h.DB.Create(&im).Error; err != nil {
		os.Remove(dst)
		return fmt.Errorf("写入数据库失败")
	}
	return nil
}

func (h *ScoreHandler) nextSortOrder(scoreID uint) int {
	var max int
	h.DB.Model(&models.ScoreImage{}).Where("score_id = ?", scoreID).
		Select("COALESCE(MAX(sort_order), -1)").Scan(&max)
	return max + 1
}

func (h *ScoreHandler) removeLocalFile(stored string) {
	if stored == "" {
		return
	}
	p := filepath.Join(h.Dir, filepath.Base(stored))
	if strings.HasPrefix(filepath.Clean(p), filepath.Clean(h.Dir)) {
		os.Remove(p)
	}
}

/* ------------------------------------------------------------------ */
/* helpers                                                             */
/* ------------------------------------------------------------------ */

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return 0, false
	}
	return id, true
}

// randomScoreName builds an unguessable stored filename: 16 hex chars + extension.
func randomScoreName(ext string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("score-%d%s", os.Getpid(), ext)
	}
	return hex.EncodeToString(b) + ext
}

