package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DenyPatterns contains regex patterns for dangerous commands.
var DenyPatterns = []string{
	`\brm\s+(-[rf]+\s+)*[/~]`, // rm with root or home
	`\brm\s+-rf\b`,            // rm -rf anywhere
	`\bdd\b.*\bof=/dev/`,      // dd to device
	`\bmkfs\b`,                // filesystem format
	`\bfdisk\b`,               // partition tool
	`\bformat\b`,              // Windows format
	`>\s*/dev/`,               // redirect to device
	`\bchmod\s+-R\s+777\b`,    // chmod 777 recursive
	`\bchown\s+-R\b.*[/~]`,    // chown recursive on root/home
	`\b:(){ :|:& };:\b`,       // fork bomb
	`\bshutdown\b`,            // shutdown
	`\breboot\b`,              // reboot
	`\bhalt\b`,                // halt
	`\binit\s+[0-6]\b`,        // init level change
	`\bsystemctl\s+(start|stop|restart|enable|disable)\b`, // systemd control
}

// PathPatterns for detecting path traversal attempts.
var PathPatterns = []string{
	`\.\.\/`, // ../
	`\.\.\\`, // ..\
	`\/\.\.`, // /..
	`\\\.\.`, // \..
}

// ExecTool executes shell commands.
type ExecTool struct {
	Timeout             time.Duration
	RestrictToWorkspace bool
	WorkDir             string
	denyRegexes         []*regexp.Regexp
	pathRegexes         []*regexp.Regexp
}

// NewExecTool creates a new ExecTool.
func NewExecTool(timeout time.Duration, restrictToWorkspace bool, workDir string) *ExecTool {
	// Compile deny patterns
	denyRegexes := make([]*regexp.Regexp, 0, len(DenyPatterns))
	for _, pattern := range DenyPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			denyRegexes = append(denyRegexes, re)
		}
	}

	// Compile path patterns
	pathRegexes := make([]*regexp.Regexp, 0, len(PathPatterns))
	for _, pattern := range PathPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			pathRegexes = append(pathRegexes, re)
		}
	}

	return &ExecTool{
		Timeout:             timeout,
		RestrictToWorkspace: restrictToWorkspace,
		WorkDir:             workDir,
		denyRegexes:         denyRegexes,
		pathRegexes:         pathRegexes,
	}
}

func (t *ExecTool) Name() string { return "exec" }

func (t *ExecTool) Description() string {
	return "Execute a shell command and return its output."
}

func (t *ExecTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]any{
				"type":        "string",
				"description": "Optional working directory for the command",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecTool) Execute(ctx context.Context, params map[string]any) (string, error) {
	command := GetString(params, "command", "")
	workingDir := GetString(params, "working_dir", t.WorkDir)

	if command == "" {
		return "Error: command is required", nil
	}

	// Security checks
	if err := t.guardCommand(command, workingDir); err != nil {
		return err.Error(), nil
	}

	// Create command with timeout
	timeout := t.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Build result
	var result strings.Builder
	if stdout.Len() > 0 {
		result.WriteString(stdout.String())
	}
	if stderr.Len() > 0 {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("STDERR:\n")
		result.WriteString(stderr.String())
	}

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Sprintf("Error: command timed out after %v\n%s", timeout, result.String()), nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.WriteString(fmt.Sprintf("\nExit code: %d", exitErr.ExitCode()))
		} else {
			return fmt.Sprintf("Error executing command: %v", err), nil
		}
	}

	if result.Len() == 0 {
		return "(no output)", nil
	}

	return result.String(), nil
}

func (t *ExecTool) guardCommand(command, workingDir string) error {
	// Check deny patterns
	for _, re := range t.denyRegexes {
		if re.MatchString(command) {
			return fmt.Errorf("Error: command blocked for safety: %s", re.String())
		}
	}

	// Check path traversal if workspace restricted
	if t.RestrictToWorkspace && t.WorkDir != "" {
		for _, re := range t.pathRegexes {
			if re.MatchString(command) {
				return fmt.Errorf("Error: path traversal not allowed")
			}
		}

		// Additional check: command shouldn't reference paths outside workspace
		absWorkDir, _ := filepath.Abs(t.WorkDir)
		if workingDir != "" && workingDir != t.WorkDir {
			absWorkingDir, _ := filepath.Abs(workingDir)
			if !strings.HasPrefix(absWorkingDir, absWorkDir) {
				return fmt.Errorf("Error: working directory must be within workspace")
			}
		}
	}

	return nil
}
