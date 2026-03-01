package router

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sebastiw/sidan-backend/src/config"
	"github.com/sebastiw/sidan-backend/src/fdroid"
)

type FDroidHandler struct{}

func NewFDroidHandler() FDroidHandler {
	return FDroidHandler{}
}

// uploadAPKHandler accepts a multipart APK upload, parses its metadata,
// writes a sidecar file, and regenerates the signed index.
// POST /fdroid/repo/upload
// Form field "apk": the APK file
// Optional form fields: name, summary, description, license, source_code, categories
func (fh FDroidHandler) uploadAPKHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetFDroid()

	if err := r.ParseMultipartForm(200 << 20); err != nil {
		http.Error(w, `{"error":"request too large or malformed"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("apk")
	if err != nil {
		http.Error(w, `{"error":"missing apk field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	if filepath.Ext(header.Filename) != ".apk" {
		http.Error(w, `{"error":"file must have .apk extension"}`, http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(cfg.RepoPath, filepath.Base(header.Filename))
	if err := saveUpload(file, destPath); err != nil {
		slog.Error("failed to save APK", "path", destPath, "error", err)
		http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
		return
	}

	info, err := fdroid.ParseAPK(destPath)
	if err != nil {
		slog.Error("failed to parse APK", "path", destPath, "error", err)
		http.Error(w, `{"error":"failed to parse APK"}`, http.StatusUnprocessableEntity)
		return
	}

	perms := info.Permissions
	if perms == nil {
		perms = []string{}
	}

	meta := &fdroid.APKMeta{
		PackageName:      info.PackageName,
		VersionCode:      info.VersionCode,
		VersionName:      info.VersionName,
		ApkName:          filepath.Base(header.Filename),
		Hash:             info.Hash,
		Size:             info.Size,
		MinSdkVersion:    info.MinSdkVersion,
		TargetSdkVersion: info.TargetSdkVersion,
		Permissions:      perms,
		AddedMs:          time.Now().UnixMilli(),
		AppName:          r.FormValue("name"),
		Summary:          r.FormValue("summary"),
		Description:      r.FormValue("description"),
		License:          r.FormValue("license"),
		SourceCode:       r.FormValue("source_code"),
		Categories:       r.FormValue("categories"),
	}

	if err := fdroid.WriteSidecar(meta, destPath); err != nil {
		slog.Error("failed to write sidecar", "error", err)
		http.Error(w, `{"error":"failed to write metadata"}`, http.StatusInternalServerError)
		return
	}

	if err := fdroid.GenerateIndex(cfg.RepoPath, cfg); err != nil {
		slog.Error("failed to regenerate F-Droid index", "error", err)
		http.Error(w, `{"error":"failed to regenerate index"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("APK uploaded and F-Droid index regenerated",
		"package", info.PackageName,
		"version_code", info.VersionCode,
		"version_name", info.VersionName,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"package_name": info.PackageName,
		"version_code": info.VersionCode,
		"version_name": info.VersionName,
	})
}

func saveUpload(src io.Reader, destPath string) error {
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
