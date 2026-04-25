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
	cwd, _ := os.Getwd()
	// Navigate up to the project root if api is the current working directory
	if filepath.Base(cwd) == "api" {
		cwd = filepath.Dir(cwd)
	}
	// Use project root directory to avoid root permission denied issues
	return filepath.Join(cwd, "godelion", "user", userID)
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
	userID := c.Locals("user_id").(string)
	path := c.Query("path", "/")

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File upload failed"})
	}

	destPath, err := validatePath(userID, filepath.Join(path, file.Filename))
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Ensure user directory exists
	os.MkdirAll(filepath.Dir(destPath), 0755)

	if err := c.SaveFile(file, destPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save file"})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "File uploaded successfully"})
}

func ListFiles(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	path := c.Query("path", "/")

	dirPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Create user dir if not exists
	os.MkdirAll(dirPath, 0755)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read directory"})
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
	userID := c.Locals("user_id").(string)
	path := c.Query("path")

	if path == "" || path == "/" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot delete root"})
	}

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete"})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Deleted successfully"})
}

func DownloadFile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	path := c.Query("path")

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	return c.SendFile(targetPath)
}

func ReadFileContent(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	path := c.Query("path")

	targetPath, err := validatePath(userID, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found or cannot be read"})
	}

	return c.JSON(fiber.Map{
		"code": 200,
		"data": string(content),
	})
}

func CreateFolder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	path := c.Query("path", "/")
	
	type Req struct {
		Name string `json:"name"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Folder name is required"})
	}

	destPath, err := validatePath(userID, filepath.Join(path, req.Name))
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create folder"})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Folder created successfully"})
}

func MoveFile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	type Req struct {
		SourcePath string `json:"source_path"`
		TargetPath string `json:"target_path"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	srcPath, err1 := validatePath(userID, req.SourcePath)
	dstPath, err2 := validatePath(userID, req.TargetPath)
	if err1 != nil || err2 != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to move file or folder"})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Moved successfully"})
}

func ExtractArchive(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	type Req struct {
		Path string `json:"path"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	targetPath, err := validatePath(userID, req.Path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	dir := filepath.Dir(targetPath)
	var cmd *exec.Cmd

	lowerPath := strings.ToLower(req.Path)
	if strings.HasSuffix(lowerPath, ".zip") {
		cmd = exec.Command("unzip", "-o", targetPath, "-d", dir)
	} else if strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz") {
		cmd = exec.Command("tar", "-xzf", targetPath, "-C", dir)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Unsupported archive format"})
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Extraction failed",
			"details": string(output),
		})
	}

	return c.JSON(fiber.Map{"code": 200, "message": "Extracted successfully"})
}
