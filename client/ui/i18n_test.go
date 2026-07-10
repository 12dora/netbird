//go:build !(linux && 386)

package main

import "testing"

func TestTranslateChineseLocales(t *testing.T) {
	t.Parallel()

	for _, locale := range []string{"zh-CN", "zh_CN.UTF-8", "zh-Hans", "zh-TW", "ZH"} {
		if got := translate(locale, "Connect"); got != "连接" {
			t.Errorf("translate(%q, Connect) = %q, want %q", locale, got, "连接")
		}
	}
}

func TestTranslateFallsBackToEnglish(t *testing.T) {
	t.Parallel()

	if got := translate("en-US", "Connect"); got != "Connect" {
		t.Errorf("translate(en-US, Connect) = %q, want %q", got, "Connect")
	}
	if got := translate("zh-CN", "untranslated key"); got != "untranslated key" {
		t.Errorf("translate should preserve untranslated strings, got %q", got)
	}
}

func TestFormatTranslation(t *testing.T) {
	t.Parallel()

	if got := translate("zh-CN", "Install version %s"); got != "安装版本 %s" {
		t.Errorf("translate format = %q, want %q", got, "安装版本 %s")
	}
}
