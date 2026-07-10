//go:build !386

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func (s *serviceClient) setDefaultFonts() {
	if !isChineseLocale(systemLocale()) {
		return
	}

	for _, fontPath := range []string{
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJKsc-Regular.otf",
	} {
		if _, err := os.Stat(fontPath); err == nil {
			_ = os.Setenv("FYNE_FONT", fontPath)
			return
		}
	}

	log.Warn("No Noto CJK font found; Chinese UI text may not render correctly")
}
