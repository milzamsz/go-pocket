package handlers

import (
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/milzamsz/go-pocket/assets"
	"github.com/pocketbase/pocketbase/core"
)

func AssetFile() func(e *core.RequestEvent) error {
	localFS := os.DirFS("assets")
	embeddedFS := fs.FS(assets.FS)

	return func(e *core.RequestEvent) error {
		rawPath := strings.TrimSpace(e.Request.PathValue("path"))
		cleanPath := normalizeAssetPath(rawPath)
		if cleanPath == "" {
			return e.NotFoundError("asset not found", nil)
		}

		if _, err := fs.Stat(localFS, cleanPath); err == nil {
			return e.FileFS(localFS, cleanPath)
		}
		if _, err := fs.Stat(embeddedFS, cleanPath); err == nil {
			return e.FileFS(embeddedFS, cleanPath)
		}

		return e.NotFoundError("asset not found", nil)
	}
}

func normalizeAssetPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	clean := path.Clean("/" + value)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, `\`) {
		return ""
	}
	return clean
}
