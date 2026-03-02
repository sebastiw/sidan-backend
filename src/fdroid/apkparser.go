package fdroid

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

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
	PermissionsSDK23 []string // uses-permission-sdk-23
	Features         []string // informational: hardware/software features declared in manifest
	NativeCode       []string // informational: ABIs present under lib/
	Sig              string   // MD5 of signing certificate DER
	Signer           string   // SHA256 of signing certificate DER
	Hash             string   // SHA256 of the entire APK file
	Size             int64
	Icon             string // resolved icon path within the APK ZIP (may be empty or an XML adaptive icon)
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
	scanAPKZip(apkPath, &info)
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
	case "uses-permission-sdk-23":
		for _, attr := range start.Attr {
			if attr.Name.Local == "name" {
				c.info.PermissionsSDK23 = append(c.info.PermissionsSDK23, attr.Value)
			}
		}
	case "uses-feature":
		for _, attr := range start.Attr {
			// skip glEsVersion entries which have no name attribute
			if attr.Name.Local == "name" && attr.Value != "" {
				c.info.Features = append(c.info.Features, attr.Value)
			}
		}
	case "application":
		for _, attr := range start.Attr {
			if attr.Name.Local == "icon" && c.info.Icon == "" {
				c.info.Icon = attr.Value
			}
		}
	}
	return nil
}

func (c *manifestCapture) Flush() error { return nil }

// iconCandidates lists well-known icon paths for non-obfuscated APKs (highest DPI first).
var iconCandidates = []string{
	"res/mipmap-xxxhdpi/ic_launcher.png",
	"res/mipmap-xxhdpi/ic_launcher.png",
	"res/mipmap-xhdpi/ic_launcher.png",
	"res/mipmap-hdpi/ic_launcher.png",
	"res/mipmap-mdpi/ic_launcher.png",
	"res/drawable-xxxhdpi/ic_launcher.png",
	"res/drawable-xxhdpi/ic_launcher.png",
	"res/drawable-xhdpi/ic_launcher.png",
	"res/drawable-hdpi/ic_launcher.png",
	"res/drawable-mdpi/ic_launcher.png",
}

// standardIconSizes lists standard Android launcher icon dimensions in pixels,
// ordered from highest to lowest DPI preference.
var standardIconSizes = []int{192, 144, 96, 72, 48}

// ExtractIcon extracts the app icon from the APK and writes it to destPath.
// iconHint is the icon path resolved from the manifest (info.Icon); may be empty or an XML path.
func ExtractIcon(apkPath, destPath, iconHint string) error {
	zr, err := zip.OpenReader(apkPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	index := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		index[f.Name] = f
	}

	tryExtract := func(name string) bool {
		f, ok := index[name]
		if !ok {
			return false
		}
		rc, err := f.Open()
		if err != nil {
			return false
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return false
		}
		return os.WriteFile(destPath, data, 0644) == nil
	}

	// 1. Try the manifest-resolved path if it's a raster image (not an XML adaptive icon).
	lower := strings.ToLower(iconHint)
	if iconHint != "" && !strings.HasSuffix(lower, ".xml") {
		if tryExtract(iconHint) {
			return nil
		}
	}

	// 2. Try well-known paths for non-obfuscated APKs.
	for _, candidate := range iconCandidates {
		if tryExtract(candidate) {
			return nil
		}
	}

	// 3. Fallback: scan for square PNGs at standard launcher icon dimensions.
	// This handles obfuscated APKs (e.g. Flutter release builds with R8 resource shrinking)
	// where icons have short random names in a flat res/ directory.
	return extractIconByDimensions(zr.File, destPath)
}

// extractIconByDimensions scans ZIP entries for PNGs whose dimensions match standard
// Android launcher icon sizes, preferring the highest DPI and largest file at each size.
func extractIconByDimensions(files []*zip.File, destPath string) error {
	type candidate struct {
		f        *zip.File
		fileSize uint64
	}
	best := make(map[int]candidate)

	for _, f := range files {
		name := f.Name
		if !strings.HasPrefix(name, "res/") || strings.HasSuffix(name, ".9.png") {
			continue
		}
		if !strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".webp") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		hdr := make([]byte, 24)
		n, _ := io.ReadFull(rc, hdr)
		rc.Close()
		if n < 24 || string(hdr[:8]) != "\x89PNG\r\n\x1a\n" {
			continue
		}

		w := int(binary.BigEndian.Uint32(hdr[16:20]))
		h := int(binary.BigEndian.Uint32(hdr[20:24]))
		if w != h {
			continue
		}
		for _, size := range standardIconSizes {
			if w == size {
				if c, exists := best[size]; !exists || f.UncompressedSize64 > c.fileSize {
					best[size] = candidate{f, f.UncompressedSize64}
				}
				break
			}
		}
	}

	for _, size := range standardIconSizes {
		c, ok := best[size]
		if !ok {
			continue
		}
		rc, err := c.f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		return os.WriteFile(destPath, data, 0644)
	}
	return errors.New("no icon found in APK")
}

// scanAPKZip scans the APK ZIP for native code ABIs and the signing certificate.
// Errors are silently ignored — these fields are best-effort.
func scanAPKZip(apkPath string, info *APKInfo) {
	zr, err := zip.OpenReader(apkPath)
	if err != nil {
		return
	}
	defer zr.Close()

	abis := make(map[string]bool)
	for _, f := range zr.File {
		// lib/<abi>/libfoo.so → collect ABI name
		if strings.HasPrefix(f.Name, "lib/") {
			if parts := strings.SplitN(f.Name, "/", 3); len(parts) == 3 && parts[1] != "" {
				abis[parts[1]] = true
			}
		}

		// META-INF/*.RSA|DSA|EC → signing certificate
		if info.Sig == "" && strings.HasPrefix(strings.ToUpper(f.Name), "META-INF/") {
			upper := strings.ToUpper(f.Name)
			if strings.HasSuffix(upper, ".RSA") || strings.HasSuffix(upper, ".DSA") || strings.HasSuffix(upper, ".EC") {
				if certDER, err := readSigningCert(f); err == nil {
					md5sum := md5.Sum(certDER)
					sha256sum := sha256.Sum256(certDER)
					info.Sig = hex.EncodeToString(md5sum[:])
					info.Signer = hex.EncodeToString(sha256sum[:])
				}
			}
		}
	}

	if len(abis) > 0 {
		info.NativeCode = make([]string, 0, len(abis))
		for abi := range abis {
			info.NativeCode = append(info.NativeCode, abi)
		}
		sort.Strings(info.NativeCode)
	}
}

func readSigningCert(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return nil, err
	}
	return extractCertFromPKCS7(data)
}

// extractCertFromPKCS7 navigates a PKCS#7 SignedData structure and returns
// the DER bytes of the first signing certificate.
func extractCertFromPKCS7(data []byte) ([]byte, error) {
	// ContentInfo { OID, [0] EXPLICIT content }
	var outer struct {
		OID     asn1.ObjectIdentifier
		Content asn1.RawValue `asn1:"explicit,tag:0"`
	}
	if _, err := asn1.Unmarshal(data, &outer); err != nil {
		return nil, err
	}

	// SignedData { version, digestAlgorithms, encapContentInfo, [0] certificates, ... }
	var sd struct {
		Version          int
		DigestAlgorithms asn1.RawValue `asn1:"set"`
		ContentInfo      asn1.RawValue
		Certificates     asn1.RawValue `asn1:"optional,tag:0"`
	}
	if _, err := asn1.Unmarshal(outer.Content.Bytes, &sd); err != nil {
		return nil, err
	}
	if len(sd.Certificates.Bytes) == 0 {
		return nil, errors.New("no certificates in PKCS#7")
	}

	// Certificates is a SET OF Certificate; grab the first one
	var cert asn1.RawValue
	if _, err := asn1.Unmarshal(sd.Certificates.Bytes, &cert); err != nil {
		return nil, err
	}
	return cert.FullBytes, nil
}

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
