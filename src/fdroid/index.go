package fdroid

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/sebastiw/sidan-backend/src/config"
)

// GenerateIndex builds index-v1.json and signs it into index-v1.jar.
// It reads all *.apk.meta.json sidecars from repoPath.
func GenerateIndex(repoPath string, cfg *config.FDroidConfiguration) error {
	metas, err := LoadAllMeta(repoPath)
	if err != nil {
		return fmt.Errorf("loading meta: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(buildIndex(metas, cfg), "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling index: %w", err)
	}

	jsonPath := filepath.Join(repoPath, "index-v1.json")
	if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("writing index-v1.json: %w", err)
	}

	return signIndex(repoPath, cfg)
}

// --- index-v1.json types ---

type fdroidIndex struct {
	Repo     repoMeta                 `json:"repo"`
	Requests requests                 `json:"requests"`
	Apps     []appMeta                `json:"apps"`
	Packages map[string][]packageMeta `json:"packages"`
}

type repoMeta struct {
	Timestamp   int64    `json:"timestamp"`
	Version     int      `json:"version"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Address     string   `json:"address"`
	Description string   `json:"description"`
	Mirrors     []string `json:"mirrors"`
}

type requests struct {
	Install   []string `json:"install"`
	Uninstall []string `json:"uninstall"`
}

type appMeta struct {
	PackageName          string   `json:"packageName"`
	Name                 string   `json:"name"`
	Summary              string   `json:"summary"`
	Description          string   `json:"description"`
	Categories           []string `json:"categories"`
	License              string   `json:"license"`
	SourceCode           string   `json:"sourceCode"`
	Added                int64    `json:"added"`
	LastUpdated          int64    `json:"lastUpdated"`
	SuggestedVersionCode int64    `json:"suggestedVersionCode"`
	SuggestedVersionName string   `json:"suggestedVersionName"`
	Icon                 string   `json:"icon"`
}

type packageMeta struct {
	VersionCode      int64    `json:"versionCode"`
	VersionName      string   `json:"versionName"`
	ApkName          string   `json:"apkName"`
	Hash             string   `json:"hash"`
	HashType         string   `json:"hashType"`
	Size             int64    `json:"size"`
	MinSdkVersion    int      `json:"minSdkVersion"`
	TargetSdkVersion int      `json:"targetSdkVersion"`
	Permissions      []string `json:"uses-permission"`
	Added            int64    `json:"added"`
}

// --- index building ---

func buildIndex(metas []APKMeta, cfg *config.FDroidConfiguration) fdroidIndex {
	byPkg := make(map[string][]APKMeta)
	for _, m := range metas {
		byPkg[m.PackageName] = append(byPkg[m.PackageName], m)
	}

	idx := fdroidIndex{
		Repo: repoMeta{
			Timestamp:   time.Now().UnixMilli(),
			Version:     20002,
			Name:        cfg.RepoName,
			Icon:        "",
			Address:     cfg.RepoAddress,
			Description: cfg.RepoDescription,
			Mirrors:     []string{},
		},
		Requests: requests{Install: []string{}, Uninstall: []string{}},
		Apps:     []appMeta{},
		Packages: make(map[string][]packageMeta),
	}

	for pkgName, versions := range byPkg {
		// Sort descending by version code so latest is first
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].VersionCode > versions[j].VersionCode
		})
		latest := versions[0]

		idx.Apps = append(idx.Apps, appMeta{
			PackageName:          pkgName,
			Name:                 coalesce(latest.AppName, pkgName),
			Summary:              latest.Summary,
			Description:          latest.Description,
			Categories:           categoriesSlice(latest.Categories),
			License:              latest.License,
			SourceCode:           latest.SourceCode,
			Added:                versions[len(versions)-1].AddedMs, // oldest upload
			LastUpdated:          latest.AddedMs,
			SuggestedVersionCode: latest.VersionCode,
			SuggestedVersionName: latest.VersionName,
			Icon:                 "icons/" + pkgName + ".png",
		})

		pkgs := make([]packageMeta, 0, len(versions))
		for _, v := range versions {
			perms := v.Permissions
			if perms == nil {
				perms = []string{}
			}
			pkgs = append(pkgs, packageMeta{
				VersionCode:      v.VersionCode,
				VersionName:      v.VersionName,
				ApkName:          v.ApkName,
				Hash:             v.Hash,
				HashType:         "sha256",
				Size:             v.Size,
				MinSdkVersion:    v.MinSdkVersion,
				TargetSdkVersion: v.TargetSdkVersion,
				Permissions:      perms,
				Added:            v.AddedMs,
			})
		}
		idx.Packages[pkgName] = pkgs
	}

	// Deterministic ordering
	sort.Slice(idx.Apps, func(i, j int) bool {
		return idx.Apps[i].PackageName < idx.Apps[j].PackageName
	})

	return idx
}

// --- JAR creation and signing ---

func signIndex(repoPath string, cfg *config.FDroidConfiguration) error {
	jarPath := filepath.Join(repoPath, "index-v1.jar")

	if err := buildJar(jarPath, repoPath); err != nil {
		return fmt.Errorf("building jar: %w", err)
	}

	cmd := exec.Command(
		"jarsigner",
		"-keystore", cfg.KeystorePath,
		"-storepass", cfg.KeystorePassword,
		"-keypass", cfg.KeyPassword,
		"-sigalg", "SHA1withRSA",
		"-digestalg", "SHA1",
		jarPath,
		cfg.KeyAlias,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("jarsigner: %w\n%s", err, out)
	}
	return nil
}

// buildJar creates an unsigned JAR (ZIP) containing index-v1.json.
// jarsigner will add the signature entries in place.
func buildJar(jarPath, repoPath string) error {
	f, err := os.Create(jarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)

	mf, err := w.Create("META-INF/MANIFEST.MF")
	if err != nil {
		w.Close()
		return err
	}
	mf.Write([]byte("Manifest-Version: 1.0\r\nCreated-By: sidan-backend\r\n\r\n"))

	jsonData, err := os.ReadFile(filepath.Join(repoPath, "index-v1.json"))
	if err != nil {
		w.Close()
		return err
	}
	jf, err := w.Create("index-v1.json")
	if err != nil {
		w.Close()
		return err
	}
	jf.Write(jsonData)

	return w.Close()
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
