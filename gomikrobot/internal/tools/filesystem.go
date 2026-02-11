package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool reads the contents of a file.
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file at the specified path."
}

func (t *ReadFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path", "")
	if path == "" {
		return "Error: path is required", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Error: file not found: %s", path), nil
		}
		if os.IsPermission(err) {
			return fmt.Sprintf("Error: permission denied: %s", path), nil
		}
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	return string(content), nil
}

// WriteFileTool writes content to a file.
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }

func (t *WriteFileTool) Description() string {
	return "Write content to a file at the specified path. Creates parent directories if needed."
}

func (t *WriteFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path", "")
	content := GetString(params, "content", "")

	if path == "" {
		return "Error: path is required", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err), nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		if os.IsPermission(err) {
			return fmt.Sprintf("Error: permission denied: %s", path), nil
		}
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// EditFileTool replaces text in a file.
type EditFileTool struct{}

func (t *EditFileTool) Name() string { return "edit_file" }

func (t *EditFileTool) Description() string {
	return "Edit a file by replacing text. Useful for making targeted changes."
}

func (t *EditFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The path to the file to edit",
			},
			"old_text": map[string]any{
				"type":        "string",
				"description": "The text to find and replace",
			},
			"new_text": map[string]any{
				"type":        "string",
				"description": "The replacement text",
			},
		},
		"required": []string{"path", "old_text", "new_text"},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path", "")
	oldText := GetString(params, "old_text", "")
	newText := GetString(params, "new_text", "")

	if path == "" {
		return "Error: path is required", nil
	}
	if oldText == "" {
		return "Error: old_text is required", nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Error: file not found: %s", path), nil
		}
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, oldText) {
		return fmt.Sprintf("Error: text not found in file: %s", path), nil
	}

	newContent := strings.Replace(contentStr, oldText, newText, 1)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Successfully edited %s", path), nil
}

// ListDirTool lists directory contents.
type ListDirTool struct{}

func (t *ListDirTool) Name() string { return "list_dir" }

func (t *ListDirTool) Description() string {
	return "List the contents of a directory."
}

func (t *ListDirTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The directory path to list",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	path := GetString(params, "path", ".")

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Error: directory not found: %s", path), nil
		}
		if os.IsPermission(err) {
			return fmt.Sprintf("Error: permission denied: %s", path), nil
		}
		return fmt.Sprintf("Error reading directory: %v", err), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of %s:\n", path))

	for _, entry := range entries {
		info, _ := entry.Info()
		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("  [DIR]  %s/\n", entry.Name()))
		} else if info != nil {
			result.WriteString(fmt.Sprintf("  [FILE] %s (%d bytes)\n", entry.Name(), info.Size()))
		} else {
			result.WriteString(fmt.Sprintf("  [FILE] %s\n", entry.Name()))
		}
	}

	return result.String(), nil
}

// NewReadFileTool creates a new ReadFileTool.
func NewReadFileTool() *ReadFileTool { return &ReadFileTool{} }

// NewWriteFileTool creates a new WriteFileTool.
func NewWriteFileTool() *WriteFileTool { return &WriteFileTool{} }

// NewEditFileTool creates a new EditFileTool.
func NewEditFileTool() *EditFileTool { return &EditFileTool{} }

// NewListDirTool creates a new ListDirTool.
func NewListDirTool() *ListDirTool { return &ListDirTool{} }
