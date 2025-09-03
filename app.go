package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the app starts up
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// CompressRequest 压缩请求结构
type CompressRequest struct {
	InputDir     string `json:"inputDir"`
	OutputDir    string `json:"outputDir"`
	Format       string `json:"format"`
	Quality      int    `json:"quality"`
	MaxWidth     int    `json:"maxWidth"`
	MaxHeight    int    `json:"maxHeight"`
}

// CompressResult 压缩结果结构
type CompressResult struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	ProcessedCount int      `json:"processedCount"`
	TotalCount     int      `json:"totalCount"`
	Errors         []string `json:"errors"`
	OriginalSize   int64    `json:"originalSize"`
	CompressedSize int64    `json:"compressedSize"`
}

// SelectFolder 选择文件夹
func (a *App) SelectFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("应用上下文未初始化")
	}
	
	result, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择文件夹",
	})
	
	return result, err
}

// TestFunction 测试函数
func (a *App) TestFunction() string {
	return "后端连接成功！"
}

// GetSupportedFormats 获取支持的格式
func (a *App) GetSupportedFormats() []string {
	return []string{"webp", "jpeg", "jpg", "png"}
}

// CompressImages 压缩图片
func (a *App) CompressImages(req CompressRequest) CompressResult {
	result := CompressResult{
		Success: false,
		Errors:  make([]string, 0),
	}

	// 检查输入目录
	if _, err := os.Stat(req.InputDir); os.IsNotExist(err) {
		result.Message = "输入目录不存在"
		return result
	}

	// 创建输出目录
	if err := os.MkdirAll(req.OutputDir, 0755); err != nil {
		result.Message = fmt.Sprintf("创建输出目录失败: %v", err)
		return result
	}

	// 获取所有图片文件
	imageFiles, err := getImageFiles(req.InputDir)
	if err != nil {
		result.Message = fmt.Sprintf("扫描图片文件失败: %v", err)
		return result
	}

	result.TotalCount = len(imageFiles)
	if result.TotalCount == 0 {
		result.Message = "未找到支持的图片文件"
		return result
	}

	// 使用协程池进行并发压缩
	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(imageFiles))
	results := make(chan compressJobResult, len(imageFiles))

	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for inputPath := range jobs {
				jobResult := compressImage(inputPath, req)
				results <- jobResult
			}
		}()
	}

	// 发送任务
	go func() {
		for _, file := range imageFiles {
			jobs <- file
		}
		close(jobs)
	}()

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	for jobResult := range results {
		if jobResult.Success {
			result.ProcessedCount++
			result.OriginalSize += jobResult.OriginalSize
			result.CompressedSize += jobResult.CompressedSize
		} else {
			result.Errors = append(result.Errors, jobResult.Error)
		}
	}

	if result.ProcessedCount > 0 {
		result.Success = true
		compressionRatio := float64(result.CompressedSize) / float64(result.OriginalSize) * 100
		result.Message = fmt.Sprintf("成功压缩 %d/%d 张图片，压缩率: %.1f%%",
			result.ProcessedCount, result.TotalCount, compressionRatio)
	} else {
		result.Message = "没有图片被成功压缩"
	}

	return result
}

type compressJobResult struct {
	Success        bool
	Error          string
	OriginalSize   int64
	CompressedSize int64
}

// compressImage 压缩单张图片
func compressImage(inputPath string, req CompressRequest) compressJobResult {
	result := compressJobResult{}

	// 获取原文件信息
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		result.Error = fmt.Sprintf("获取文件信息失败 %s: %v", inputPath, err)
		return result
	}
	result.OriginalSize = fileInfo.Size()

	// 打开图片
	src, err := imaging.Open(inputPath)
	if err != nil {
		result.Error = fmt.Sprintf("打开图片失败 %s: %v", inputPath, err)
		return result
	}

	// 如果设置了最大尺寸，进行缩放
	if req.MaxWidth > 0 || req.MaxHeight > 0 {
		bounds := src.Bounds()
		width := bounds.Max.X
		height := bounds.Max.Y

		if (req.MaxWidth > 0 && width > req.MaxWidth) || (req.MaxHeight > 0 && height > req.MaxHeight) {
			src = imaging.Fit(src, req.MaxWidth, req.MaxHeight, imaging.Lanczos)
		}
	}

	// 生成输出文件路径
	filename := filepath.Base(inputPath)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	outputPath := filepath.Join(req.OutputDir, nameWithoutExt+"."+req.Format)

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		result.Error = fmt.Sprintf("创建输出文件失败 %s: %v", outputPath, err)
		return result
	}
	defer outputFile.Close()

	// 根据格式进行压缩
	switch strings.ToLower(req.Format) {
	case "webp":
		err = webp.Encode(outputFile, src, &webp.Options{
			Lossless: false,
			Quality:  float32(req.Quality),
		})
	case "jpeg", "jpg":
		err = jpeg.Encode(outputFile, src, &jpeg.Options{
			Quality: req.Quality,
		})
	case "png":
		err = png.Encode(outputFile, src)
	default:
		result.Error = fmt.Sprintf("不支持的格式: %s", req.Format)
		return result
	}

	if err != nil {
		result.Error = fmt.Sprintf("编码图片失败 %s: %v", outputPath, err)
		return result
	}

	// 获取压缩后文件大小
	if outputInfo, err := os.Stat(outputPath); err == nil {
		result.CompressedSize = outputInfo.Size()
	}

	result.Success = true
	return result
}

// getImageFiles 获取目录下的所有图片文件
func getImageFiles(dir string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".bmp":  true,
		".tiff": true,
		".gif":  true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
