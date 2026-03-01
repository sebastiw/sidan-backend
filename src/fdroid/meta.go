package fdroid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// APKMeta is the sidecar file stored alongside each APK on disk.
// It contains everything needed to build the F-Droid index for that APK version.
type APKMeta struct {
	// Parsed from the APK itself
	PackageName      string   `json:"package_name"`
	VersionCode      int64    `json:"version_code"`
	VersionName      string   `json:"version_name"`
	ApkName          string   `json:"apk_name"`
	Hash             string   `json:"hash"`
	Size             int64    `json:"size"`
	MinSdkVersion    int      `json:"min_sdk_version"`
	TargetSdkVersion int      `json:"target_sdk_version"`
	Permissions      []string `json:"permissions"`
	AddedMs          int64    `json:"added_ms"` // unix milliseconds at upload time

	// Supplementary — provided optionally at upload time, stored as-is
	AppName     string `json:"app_name"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	License     string `json:"license"`
	SourceCode  string `json:"source_code"`
	Categories  string `json:"categories"` // comma-separated
}

// SidecarPath returns the sidecar path for a given APK path.
func SidecarPath(apkPath string) string {
	return apkPath + ".meta.json"
}

// WriteSidecar serialises meta to its sidecar file next to the APK.
func WriteSidecar(meta *APKMeta, apkPath string) error {
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(SidecarPath(apkPath), b, 0644)
}

// LoadAllMeta reads every *.apk.meta.json sidecar in repoPath.
func LoadAllMeta(repoPath string) ([]APKMeta, error) {
	matches, err := filepath.Glob(filepath.Join(repoPath, "*.apk.meta.json"))
	if err != nil {
		return nil, err
	}

	var metas []APKMeta
	for _, path := range matches {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var m APKMeta
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return metas, nil
}

// categoriesSlice splits the comma-separated categories string into a slice.
func categoriesSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
