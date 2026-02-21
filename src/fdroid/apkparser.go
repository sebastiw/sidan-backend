package fdroid

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"io"
	"os"
	"strconv"

	"github.com/avast/apkparser"
)

// APKInfo contains metadata extracted from an APK file.
type APKInfo struct {
	PackageName      string
	VersionCode      int64
	VersionName      string
	MinSdkVersion    int
	TargetSdkVersion int
	Permissions      []string
	Hash             string // SHA256 hex of the entire APK file
	Size             int64
}

// ParseAPK extracts metadata from an APK file at the given path.
func ParseAPK(apkPath string) (*APKInfo, error) {
	hash, size, err := sha256File(apkPath)
	if err != nil {
		return nil, err
	}

	enc := &manifestCapture{}
	zipErr, _, manifestErr := apkparser.ParseApk(apkPath, enc)
	if zipErr != nil {
		return nil, zipErr
	}
	if manifestErr != nil {
		return nil, manifestErr
	}

	info := enc.info
	info.Hash = hash
	info.Size = size
	return &info, nil
}

// manifestCapture implements apkparser.ManifestEncoder and extracts the fields we need.
type manifestCapture struct {
	info APKInfo
}

func (c *manifestCapture) EncodeToken(t xml.Token) error {
	start, ok := t.(xml.StartElement)
	if !ok {
		return nil
	}

	switch start.Name.Local {
	case "manifest":
		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "package":
				c.info.PackageName = attr.Value
			case "versionCode":
				c.info.VersionCode, _ = strconv.ParseInt(attr.Value, 10, 64)
			case "versionName":
				c.info.VersionName = attr.Value
			}
		}
	case "uses-sdk":
		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "minSdkVersion":
				c.info.MinSdkVersion, _ = strconv.Atoi(attr.Value)
			case "targetSdkVersion":
				c.info.TargetSdkVersion, _ = strconv.Atoi(attr.Value)
			}
		}
	case "uses-permission":
		for _, attr := range start.Attr {
			if attr.Name.Local == "name" {
				c.info.Permissions = append(c.info.Permissions, attr.Value)
			}
		}
	}
	return nil
}

func (c *manifestCapture) Flush() error { return nil }

func sha256File(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), size, nil
}
