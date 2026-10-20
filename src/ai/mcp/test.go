package mcp

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"hatch/core/config"
	"hatch/core/macos"
)

// TestConnection runs `claude mcp get <id>` (live connection check) or
// `codex mcp get <id>` (config check only) and summarizes the result. The
// caller stamps Target/CheckedAt and persists it.
func TestConnection(id, target string) config.MCPTestResult {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	if target == TargetClaude {
		if macos.FindCLI("claude") == "" {
			return TestFailure(target, "claude CLI 未安装或不在 PATH", "", -1, "connection")
		}
		cmd = ExecCommand(ctx, "claude", "mcp", "get", id)
	} else {
		if macos.FindCLI("codex") == "" {
			return TestFailure(target, "codex CLI 未安装或不在 PATH", "", -1, "config")
		}
		cmd = ExecCommand(ctx, "codex", "mcp", "get", id)
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	detail := strings.TrimSpace(buf.String())
	exitCode := 0
	if err != nil {
		exitCode = commandExitCode(err)
		if ctx.Err() == context.DeadlineExceeded {
			exitCode = -1
			detail = strings.TrimSpace(detail + "\nconnection test timed out")
		}
	}
	return ParseTestOutput(target, detail, exitCode, err)
}

// ParseTestOutput turns `claude|codex mcp get <id>` output into a test result.
func ParseTestOutput(target, detail string, exitCode int, err error) config.MCPTestResult {
	mode := "connection"
	if target == TargetCodex {
		mode = "config"
	}
	if err != nil {
		return TestFailure(target, summarizeFailure(detail), detail, exitCode, mode)
	}
	if target == TargetClaude {
		if strings.Contains(detail, "Status: ✓") || strings.Contains(detail, "Connected") {
			return config.MCPTestResult{Target: target, Status: "success", Summary: "Claude Code MCP 连接成功", Detail: detail, Mode: mode}
		}
		if line := firstLineContaining(detail, "Status:"); line != "" {
			return TestFailure(target, strings.TrimSpace(strings.TrimPrefix(line, "Status:")), detail, exitCode, mode)
		}
		return config.MCPTestResult{Target: target, Status: "unknown", Summary: "Claude Code 返回了配置，但未包含连接状态", Detail: detail, Mode: mode}
	}
	if strings.Contains(detail, "enabled: true") || strings.Contains(detail, "Status") || strings.Contains(detail, "transport:") {
		return config.MCPTestResult{Target: target, Status: "unknown", Summary: "Codex 配置可读取；当前 CLI 未提供实时连接状态", Detail: detail, Mode: mode}
	}
	return config.MCPTestResult{Target: target, Status: "unknown", Summary: "Codex MCP 配置检查完成", Detail: detail, Mode: mode}
}

// TestFailure builds a failed test result.
func TestFailure(target, summary, detail string, exitCode int, mode string) config.MCPTestResult {
	return config.MCPTestResult{Target: target, Status: "failed", Summary: summary, Detail: detail, ExitCode: exitCode, Mode: mode}
}

func summarizeFailure(detail string) string {
	detail = strings.TrimSpace(detail)
	if detail == "" {
		return "连接测试失败，但 CLI 未返回错误详情"
	}
	for _, line := range strings.Split(detail, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "连接测试失败"
}

func firstLineContaining(text, needle string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}
