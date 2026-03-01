# Android (F-Droid) Repository Implementation Plan

## What is an F-Droid Repository

F-Droid is an Android FOSS app store. Any server can act as a compatible repository by
serving a specific set of files. The F-Droid Android client is pointed at a base URL, fetches
a signed index, and lets users install/update apps directly from the repo.

This is a **private, first-party repo**. Only we upload APKs. No third-party submissions.

### Files the F-Droid client needs at `{baseURL}/fdroid/repo/`

| File | Description |
|---|---|
| `index-v1.jar` | **Required.** Signed JAR containing `index-v1.json`. Client verifies the signature. |
| `index-v1.json` | JSON index with all app and APK metadata (also inside the JAR). |
| `*.apk` | The actual APK files, served statically. |

### The signing constraint

The JAR signature must be made with a stable RSA keystore. There is no Go library that
produces a JAR signature F-Droid clients accept. The plan shells out to `jarsigner` (part of
the JDK). This is a one-time setup and a deployment requirement.

---

## Approach: No Database

Since we control all APKs and there are only a handful, there is no need for a database table.
Instead:

- On APK upload, the backend parses the APK and writes a `.meta.json` sidecar file next to the APK on disk.
- Index generation scans all `*.meta.json` files in the repo directory and builds `index-v1.json` from them.
- Additional app metadata (description, license, etc.) that is not inside the APK can be
  provided as optional form fields at upload time and stored in the sidecar.

The repo directory ends up looking like this:

```
static/fdroid/repo/
├── index-v1.json                          (generated)
├── index-v1.jar                           (generated + signed)
├── com.example.myapp_42.apk
├── com.example.myapp_42.apk.meta.json     (sidecar: parsed + supplemental metadata)
├── com.example.myapp_43.apk
├── com.example.myapp_43.apk.meta.json
└── icons/
    └── com.example.myapp.png              (extracted from APK on first upload)
```

The sidecar is the only persistent store. The index is always regenerated from the sidecars.

---

## Step 1: Prerequisites and Configuration

### 1a. JDK on PATH

`jarsigner` and `jar` must be available on the PATH where the server runs. Document this as a
deployment requirement. In Docker, add to the image:

```dockerfile
RUN apt-get install -y default-jdk-headless
```

### 1b. Generate a keystore (one-time, done by the admin)

```bash
keytool -genkey -v \
  -keystore fdroid.keystore \
  -alias fdroid \
  -keyalg RSA -keysize 4096 -validity 10000
```

**The keystore must never change.** Changing the signing key will break the repo for anyone
who already added it.

### 1c. New config block in `config/local.yaml`

```yaml
fdroid:
  repoName: "Sidan Apps"
  repoDescription: "Chalmers Losers official app repository"
  repoAddress: "https://api.chalmerslosers.com/fdroid/repo"
  repoPath: "./static/fdroid/repo"
  keystorePath: "./fdroid.keystore"
  keystorePassword: "changeme"
  keyAlias: "fdroid"
  keyPassword: "changeme"
```

### 1d. New config struct in `src/config/config.go`

```go
type FDroidConfiguration struct {
    RepoName         string
    RepoDescription  string
    RepoAddress      string
    RepoPath         string
    KeystorePath     string
    KeystorePassword string
    KeyAlias         string
    KeyPassword      string
}
```

Add `FDroid FDroidConfiguration` to `Configuration`, add Viper bindings, and add a
`GetFDroid() *FDroidConfiguration` accessor.

---

## Step 2: Auth Scope

Add `WriteFDroidScope = "write:apk"` to the scope constants in `src/auth/middleware.go`.

Assign this scope in `src/auth/jwt.go` when generating tokens for members who should be
allowed to upload (e.g. member type `#`, i.e. full members).

---

## Step 3: APK Parser Package

Create `src/fdroid/apkparser.go`.

**New dependency**: add `github.com/avast/apkparser` to `go.mod`. It handles the binary
`AndroidManifest.xml` format inside APKs (a plain `archive/zip` cannot read it).

### What to extract

```go
type APKInfo struct {
    PackageName      string
    VersionCode      int64
    VersionName      string
    MinSdkVersion    int
    TargetSdkVersion int
    Permissions      []string // e.g. ["android.permission.INTERNET"]
    SignerFingerprint string   // SHA256 hex of the signing cert
    Icon             []byte   // PNG bytes of the launcher icon, may be nil
    Hash             string   // SHA256 hex of the whole APK file
    Size             int64
}

func ParseAPK(apkPath string) (*APKInfo, error)
```

Implementation:
- Open as `archive/zip`
- Parse `AndroidManifest.xml` with `apkparser`
- Collect `<uses-permission>` elements
- For the signing cert: read `META-INF/*.RSA`, parse PKCS7 with `crypto/x509`, SHA256 the DER bytes
- For the icon: resolve the icon resource via `resources.arsc`, pick highest-density PNG from the zip
- SHA256 the whole file for `Hash`

---

## Step 4: Sidecar Metadata Format

The sidecar file (`{apkName}.meta.json`) stores everything needed to build the index for one
APK version, plus optional supplementary metadata the uploader provides:

```go
// src/fdroid/meta.go
type APKMeta struct {
    // Parsed from APK
    PackageName      string   `json:"package_name"`
    VersionCode      int64    `json:"version_code"`
    VersionName      string   `json:"version_name"`
    ApkName          string   `json:"apk_name"`
    Hash             string   `json:"hash"`
    Size             int64    `json:"size"`
    MinSdkVersion    int      `json:"min_sdk_version"`
    TargetSdkVersion int      `json:"target_sdk_version"`
    SignerFingerprint string   `json:"signer"`
    Permissions      []string `json:"permissions"`
    AddedMs          int64    `json:"added_ms"` // unix ms at upload time

    // Supplementary — provided optionally at upload time, defaults to empty
    AppName     string `json:"app_name"`
    Summary     string `json:"summary"`
    Description string `json:"description"`
    License     string `json:"license"`
    SourceCode  string `json:"source_code"`
    Categories  string `json:"categories"` // comma-separated
}
```

Reading all sidecars for a repo is then just:

```go
func LoadAllMeta(repoPath string) ([]APKMeta, error) {
    // filepath.Glob(repoPath + "/*.meta.json") → unmarshal each
}
```

---

## Step 5: Index Generation and Signing

Create `src/fdroid/index.go`.

### 5a. index-v1.json structure (only the fields we need)

```go
type fdroidIndex struct {
    Repo     repoMeta                 `json:"repo"`
    Requests requests                 `json:"requests"`
    Apps     []appMeta                `json:"apps"`
    Packages map[string][]packageMeta `json:"packages"`
}
```

For each unique `PackageName` across all sidecars, emit one `appMeta` entry (using metadata
from the sidecar with the highest `VersionCode`) and a `packageMeta` list for all versions.

### 5b. `GenerateIndex(repoPath string, cfg *config.FDroidConfiguration) error`

1. `LoadAllMeta(repoPath)` — read all sidecars
2. Group by `PackageName`, sort each group by `VersionCode` descending
3. Build `fdroidIndex` struct, set `Repo.Timestamp = time.Now().UnixMilli()`
4. `json.MarshalIndent` → write to `{repoPath}/index-v1.json`
5. Call `signIndex(repoPath, cfg)` to produce `index-v1.jar`

### 5c. `signIndex(repoPath string, cfg *config.FDroidConfiguration) error`

```
# Build an unsigned JAR containing only index-v1.json
jar cf {repoPath}/index-v1-unsigned.jar -C {repoPath} index-v1.json

# Sign it in place
jarsigner \
  -keystore {keystorePath} \
  -storepass {keystorePassword} \
  -keypass {keyPassword} \
  -sigalg SHA256withRSA \
  -digestalg SHA256 \
  {repoPath}/index-v1-unsigned.jar \
  {keyAlias}

# Rename to final name
mv {repoPath}/index-v1-unsigned.jar {repoPath}/index-v1.jar
```

Use `os/exec`. Capture stderr and return it as an error if the command fails.

---

## Step 6: Upload Handler

Create `src/router/fdroid.go` with a single handler.

### `POST /fdroid/repo/upload` (requires `write:apk` scope)

1. `r.ParseMultipartForm(200 << 20)` — allow up to 200 MB
2. Get file from field `"apk"`, check extension is `.apk`
3. Save to `{repoPath}/{originalFilename}` (overwrite if same filename — i.e. same version re-upload is fine)
4. Call `fdroid.ParseAPK(savedPath)` → `APKInfo`
5. If `APKInfo.Icon != nil`, write to `{repoPath}/icons/{packageName}.png`
6. Build `APKMeta` from `APKInfo` plus optional form fields:
   `r.FormValue("name")`, `r.FormValue("summary")`, `r.FormValue("description")`,
   `r.FormValue("license")`, `r.FormValue("source_code")`, `r.FormValue("categories")`
7. Write sidecar: `json.MarshalIndent(meta)` → `{savedPath}.meta.json`
8. Call `fdroid.GenerateIndex(repoPath, config.GetFDroid())`
9. Return `{"package_name": "...", "version_code": ...}` HTTP 201

---

## Step 7: Router Registration

In `src/router/router.go`, inside `Mux()`:

```go
fdroidH := NewFDroidHandler()

// Must be registered BEFORE the file-server prefix to take precedence
r.Handle("/fdroid/repo/upload",
    authMiddleware.RequireAuth(
        authMiddleware.RequireScope(a.WriteFDroidScope)(
            http.HandlerFunc(fdroidH.uploadAPKHandler),
        ),
    ),
).Methods("POST", "OPTIONS")

// Public static serving of the entire repo directory
repoPath := config.GetFDroid().RepoPath
r.PathPrefix("/fdroid/repo/").Handler(
    http.StripPrefix("/fdroid/repo/", http.FileServer(http.Dir(repoPath))),
).Methods("GET", "HEAD")
```

---

## Step 8: Startup

In `src/sidan-backend.go`, after config is loaded, ensure the directories exist:

```go
os.MkdirAll(config.GetFDroid().RepoPath+"/icons", 0755)
```

Optionally, call `fdroid.GenerateIndex(...)` on startup to refresh the JAR after a redeploy.

---

## Summary

### New files

| File | Purpose |
|---|---|
| `src/fdroid/apkparser.go` | Parse APK binary manifest, extract metadata + icon |
| `src/fdroid/meta.go` | APKMeta struct + LoadAllMeta filesystem scan |
| `src/fdroid/index.go` | Build index-v1.json + shell out to jarsigner |
| `src/router/fdroid.go` | Upload handler |

### Modified files

| File | Change |
|---|---|
| `src/config/config.go` | Add FDroidConfiguration |
| `src/auth/middleware.go` | Add WriteFDroidScope constant |
| `src/auth/jwt.go` | Assign write:apk scope to relevant member types |
| `src/router/router.go` | Register upload + static file routes |
| `src/sidan-backend.go` | Create repo directories on startup |
| `config/local.yaml` | Add fdroid config block |
| `go.mod` | Add github.com/avast/apkparser |
| `swagger.yaml` | Document POST /fdroid/repo/upload |

**No database changes.**

---

## References

- [F-Droid API Documentation](https://f-droid.org/docs/All_our_APIs/)
- [F-Droid Setup an App Repo](https://f-droid.org/docs/Setup_an_F-Droid_App_Repo/)
- [avast/apkparser Go library](https://github.com/avast/apkparser)
