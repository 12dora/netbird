package event

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var zhCNTranslations = map[string]string{
	"Critical":  "严重",
	"Error":     "错误",
	"Warning":   "警告",
	"Info":      "信息",
	"Network":   "网络",
	"Authentication": "身份验证",
	"Connectivity":   "连接状态",
	"System":         "系统",
	" ID: %s":       " ID：%s",
}

func localize(message string) string {
	if !isChineseLocale(systemLocale()) {
		return message
	}
	if translated, ok := zhCNTranslations[message]; ok {
		return translated
	}
	return message
}

func isChineseLocale(locale string) bool {
	locale = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "_", "-"))
	return strings.HasPrefix(locale, "zh") || strings.Contains(locale, "hans") || strings.Contains(locale, "hant")
}

func systemLocale() string {
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		if locale := os.Getenv(name); locale != "" {
			return strings.Split(locale, ":")[0]
		}
	}

	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("defaults", "read", "-g", "AppleLocale")
	case "windows":
		command = exec.Command("powershell", "-NoProfile", "-Command", "(Get-Culture).Name")
	}
	if command == nil {
		return ""
	}

	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
