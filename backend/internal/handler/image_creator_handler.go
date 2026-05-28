package handler

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const imageCreatorMaxReferenceUploadBytes int64 = 25 << 20

type imageCreatorService interface {
	CreateTask(ctx context.Context, userID int64, input service.ImageCreatorCreateTaskInput) (*service.ImageCreatorTask, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]service.ImageCreatorTask, error)
	GetTask(ctx context.Context, userID int64, taskID int64) (*service.ImageCreatorTask, error)
	ListImages(ctx context.Context, userID int64, filters service.ImageCreatorImageListFilters) ([]service.ImageCreatorManagedImage, int, error)
	DeleteImages(ctx context.Context, userID int64, ids []int64) (int, error)
	GetImageFile(ctx context.Context, userID int64, imageID int64) (*service.ImageCreatorFile, error)
	GetReferenceImageForUser(ctx context.Context, userID int64, imageID int64) (*service.ImageCreatorFile, error)
}

type ImageCreatorHandler struct {
	svc imageCreatorService
}

func NewImageCreatorHandler(svc *service.ImageCreatorService) *ImageCreatorHandler {
	return &ImageCreatorHandler{svc: svc}
}

type imageCreatorCreateTaskRequest struct {
	APIKeyID     int64  `json:"api_key_id"`
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	Size         string `json:"size"`
	Quality      string `json:"quality"`
	Count        int    `json:"count"`
	OutputFormat string `json:"output_format"`
	Background   string `json:"background"`
}

type imageCreatorListResponse struct {
	Tasks  []service.ImageCreatorTask  `json:"tasks"`
	Images []service.ImageCreatorImage `json:"images"`
}

type imageCreatorImageListResponse struct {
	Items  []service.ImageCreatorManagedImage `json:"items"`
	Total  int                                `json:"total"`
	Limit  int                                `json:"limit"`
	Offset int                                `json:"offset"`
}

type imageCreatorDeleteImagesRequest struct {
	IDs []int64 `json:"ids"`
}

type imageCreatorDeleteImagesResponse struct {
	Deleted int `json:"deleted"`
}

func (h *ImageCreatorHandler) CreateTask(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	input, err := parseImageCreatorCreateTaskInput(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	task, err := h.svc.CreateTask(c.Request.Context(), subject.UserID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageCreatorHandler) ListTasks(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	tasks, err := h.svc.ListTasks(c.Request.Context(), subject.UserID, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreatorListResponse{
		Tasks:  tasks,
		Images: flattenImageCreatorImages(tasks),
	})
}

func (h *ImageCreatorHandler) ListImages(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	filters := service.ImageCreatorImageListFilters{
		Limit:       parseBoundedQueryInt(c, "limit", 40, 1, 100),
		Offset:      parseBoundedQueryInt(c, "offset", 0, 0, 100000),
		Search:      c.Query("q"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		Format:      c.Query("format"),
		Orientation: c.Query("orientation"),
		Resolution:  c.Query("resolution"),
		AspectRatio: c.Query("aspect_ratio"),
		MinWidth:    parseBoundedQueryInt(c, "min_width", 0, 0, 100000),
		MinHeight:   parseBoundedQueryInt(c, "min_height", 0, 0, 100000),
	}
	images, total, err := h.svc.ListImages(c.Request.Context(), subject.UserID, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreatorImageListResponse{
		Items:  images,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	})
}

func (h *ImageCreatorHandler) DeleteImages(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req imageCreatorDeleteImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	deleted, err := h.svc.DeleteImages(c.Request.Context(), subject.UserID, req.IDs)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreatorDeleteImagesResponse{Deleted: deleted})
}

func (h *ImageCreatorHandler) GetTask(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	taskID, ok := parsePositiveInt64Param(c, "id", "Invalid task ID")
	if !ok {
		return
	}
	task, err := h.svc.GetTask(c.Request.Context(), subject.UserID, taskID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageCreatorHandler) GetImageFile(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	imageID, ok := parsePositiveInt64Param(c, "id", "Invalid image ID")
	if !ok {
		return
	}
	file, err := h.svc.GetImageFile(c.Request.Context(), subject.UserID, imageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if file.ContentType != "" {
		c.Header("Content-Type", file.ContentType)
	}
	if strings.TrimSpace(file.FileName) != "" {
		c.Header("Content-Disposition", `inline; filename="`+strings.ReplaceAll(file.FileName, `"`, `\"`)+`"`)
	}
	serveImageCreatorFile(c, file)
}

func (h *ImageCreatorHandler) GetReferenceImageFile(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	imageID, ok := parsePositiveInt64Param(c, "id", "Invalid image ID")
	if !ok {
		return
	}
	file, err := h.svc.GetReferenceImageForUser(c.Request.Context(), subject.UserID, imageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if file.ContentType != "" {
		c.Header("Content-Type", file.ContentType)
	}
	if strings.TrimSpace(file.FileName) != "" {
		c.Header("Content-Disposition", `inline; filename="reference-`+strings.ReplaceAll(file.FileName, `"`, `\"`)+`"`)
	}
	serveImageCreatorFile(c, file)
}

func serveImageCreatorFile(c *gin.Context, file *service.ImageCreatorFile) {
	var reader io.Reader
	size := file.SizeBytes
	var modTime time.Time
	if file.Body != nil {
		defer func() { _ = file.Body.Close() }()
		reader = file.Body
	} else {
		f, err := os.Open(file.Path)
		if err != nil {
			response.NotFound(c, "image file not found")
			return
		}
		defer func() { _ = f.Close() }()
		info, err := f.Stat()
		if err != nil {
			response.NotFound(c, "image file not found")
			return
		}
		reader = f
		size = info.Size()
		modTime = info.ModTime()
	}
	if file.DownloadBytesPerSecond > 0 {
		reader = newThrottledReader(reader, file.DownloadBytesPerSecond)
	}
	c.Header("Accept-Ranges", "none")
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}
	if !modTime.IsZero() {
		c.Writer.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	}
	c.Status(http.StatusOK)
	if c.Request.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(c.Writer, reader)
}

type throttledReader struct {
	reader         io.Reader
	bytesPerSecond int64
	lastRead       time.Time
}

func newThrottledReader(reader io.Reader, bytesPerSecond int64) io.Reader {
	if bytesPerSecond <= 0 {
		return reader
	}
	return &throttledReader{reader: reader, bytesPerSecond: bytesPerSecond}
}

func (r *throttledReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n <= 0 || r.bytesPerSecond <= 0 {
		return n, err
	}
	expected := time.Duration(int64(n) * int64(time.Second) / r.bytesPerSecond)
	if !r.lastRead.IsZero() {
		if sleepFor := r.lastRead.Add(expected).Sub(time.Now()); sleepFor > 0 {
			time.Sleep(sleepFor)
		}
	}
	r.lastRead = time.Now()
	return n, err
}

func parseImageCreatorCreateTaskInput(c *gin.Context) (service.ImageCreatorCreateTaskInput, error) {
	contentType := strings.ToLower(c.ContentType())
	if strings.Contains(contentType, "multipart/form-data") {
		return parseImageCreatorMultipartInput(c)
	}
	var req imageCreatorCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return service.ImageCreatorCreateTaskInput{}, err
	}
	return imageCreatorInputFromRequest(req), nil
}

func parseImageCreatorMultipartInput(c *gin.Context) (service.ImageCreatorCreateTaskInput, error) {
	if err := c.Request.ParseMultipartForm(imageCreatorMaxReferenceUploadBytes); err != nil {
		return service.ImageCreatorCreateTaskInput{}, err
	}
	req := imageCreatorCreateTaskRequest{
		APIKeyID:     parseInt64Form(c, "api_key_id"),
		Model:        c.PostForm("model"),
		Prompt:       c.PostForm("prompt"),
		Size:         c.PostForm("size"),
		Quality:      c.PostForm("quality"),
		Count:        parseIntForm(c, "count"),
		OutputFormat: c.PostForm("output_format"),
		Background:   c.PostForm("background"),
	}
	input := imageCreatorInputFromRequest(req)
	file, header, err := c.Request.FormFile("reference_image")
	if err == http.ErrMissingFile {
		file, header, err = c.Request.FormFile("image")
	}
	if err == http.ErrMissingFile {
		return input, nil
	}
	if err != nil {
		return service.ImageCreatorCreateTaskInput{}, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, imageCreatorMaxReferenceUploadBytes+1))
	if err != nil {
		return service.ImageCreatorCreateTaskInput{}, err
	}
	if int64(len(data)) > imageCreatorMaxReferenceUploadBytes {
		return service.ImageCreatorCreateTaskInput{}, http.ErrContentLength
	}
	input.ReferenceImage = data
	input.ReferenceImageMimeType = header.Header.Get("Content-Type")
	input.ReferenceImageFilename = header.Filename
	return input, nil
}

func imageCreatorInputFromRequest(req imageCreatorCreateTaskRequest) service.ImageCreatorCreateTaskInput {
	return service.ImageCreatorCreateTaskInput{
		APIKeyID:     req.APIKeyID,
		Model:        req.Model,
		Prompt:       req.Prompt,
		Size:         req.Size,
		Quality:      req.Quality,
		Count:        req.Count,
		OutputFormat: req.OutputFormat,
		Background:   req.Background,
	}
}

func flattenImageCreatorImages(tasks []service.ImageCreatorTask) []service.ImageCreatorImage {
	images := make([]service.ImageCreatorImage, 0)
	for i := range tasks {
		images = append(images, tasks[i].Images...)
	}
	return images
}

func parsePositiveInt64Param(c *gin.Context, name string, message string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, message)
		return 0, false
	}
	return id, true
}

func parseInt64Form(c *gin.Context, name string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(c.PostForm(name)), 10, 64)
	return value
}

func parseIntForm(c *gin.Context, name string) int {
	value, _ := strconv.Atoi(strings.TrimSpace(c.PostForm(name)))
	return value
}

func parseBoundedQueryInt(c *gin.Context, name string, fallback int, minValue int, maxValue int) int {
	value := fallback
	if raw := strings.TrimSpace(c.Query(name)); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			value = parsed
		}
	}
	if value < minValue {
		return minValue
	}
	if maxValue >= minValue && value > maxValue {
		return maxValue
	}
	return value
}
