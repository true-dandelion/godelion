package controllers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func getUserDir(userID string) string {
	// Use /opt/godelion as the system-level storage root
	return filepath.Join("/opt", "godelion", "user", userID)
}

func validatePath(userID, requestedPath string) (string, error) {
	userDir := getUserDir(userID)
	fullPath := filepath.Join(userDir, requestedPath)
	
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, userDir+string(filepath.Separator)) && cleanPath != userDir {
		return "", fmt.Errorf("invalid path")
	}
	return cleanPath, nil
}

func UploadFile(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path", "/")

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "文件上传失败"})
	}

	destPath, err := validatePath(userID, filepath.Join(path, file.Filename))
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	// Ensure user directory exists
	os.MkdirAll(filepath.Dir(destPath), 0755)

	if err := c.SaveFile(file, destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "保存文件失败"})
	}

	LogAction(c, "Upload", "File", "Uploaded file: "+file.Filename+" to "+path)

	return c.JSON(fiber.Map{"code": 200, "message": "File uploaded successfully"})
}

func ListFiles(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path", "/")

	dirPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	// Create user dir if not exists
	os.MkdirAll(dirPath, 0755)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "读取目录失败"})
	}

	type FileInfo struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
		Size  int64  `json:"size"`
	}

	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Success",
		"data":    files,
	})
}

func DeleteFile(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path")

	if path == "" || path == "/" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "无法删除根目录"})
	}

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "删除失败"})
	}

	LogAction(c, "Delete", "File", "Deleted file or folder: "+path)

	return c.JSON(fiber.Map{"code": 200, "message": "Deleted successfully"})
}

func DownloadFile(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path")

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	return c.SendFile(targetPath)
}

func ReadFileContent(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path")

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "文件不存在或无法读取"})
	}

	return c.JSON(fiber.Map{
		"code": 200,
		"data": string(content),
	})
}

func CreateFolder(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))
	path := c.Query("path", "/")
	
	type Req struct {
		Name string `json:"name"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请输入文件夹名称"})
	}

	destPath, err := validatePath(userID, filepath.Join(path, req.Name))
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "创建文件夹失败"})
	}

	LogAction(c, "Create", "Folder", "Created folder: "+destPath)

	return c.JSON(fiber.Map{"code": 200, "message": "Folder created successfully"})
}

func MoveFile(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))

	type Req struct {
		SourcePath string `json:"source_path"`
		TargetPath string `json:"target_path"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	srcPath, err1 := validatePath(userID, req.SourcePath)
	dstPath, err2 := validatePath(userID, req.TargetPath)
	if err1 != nil || err2 != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "移动失败"})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Moved successfully"})
}

func ExtractArchive(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))

	type Req struct {
		Path string `json:"path"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	targetPath, err := validatePath(userID, req.Path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	dir := filepath.Dir(targetPath)
	var cmd *exec.Cmd

	lowerPath := strings.ToLower(req.Path)
	if strings.HasSuffix(lowerPath, ".zip") {
		cmd = exec.Command("unzip", "-o", targetPath, "-d", dir)
	} else if strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz") {
		cmd = exec.Command("tar", "-xzf", targetPath, "-C", dir)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "不支持的压缩格式"})
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "解压失败",
			"details": string(output),
		})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Extracted successfully"})
}

func SaveFileContent(c *fiber.Ctx) error {
	userID := fmt.Sprintf("%v", c.Locals("user_id"))

	type Req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "请求格式错误"})
	}

	targetPath, err := validatePath(userID, req.Path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "访问被拒绝"})
	}

	if err := os.WriteFile(targetPath, []byte(req.Content), 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "保存文件失败"})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "File saved successfully",
	})
}
