// Package updater handles self-updating the go-agent binary from GitHub
// Releases. A CLI can replace its own executable in place, so there is no need
// for the two-binary handoff a long-running service would require.
package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repo    = "tqrcisio/go-agent"
	binName = "go-agent"

	apiLatest = "https://api.github.com/repos/" + repo + "/releases/latest"
	dlLatest  = "https://github.com/" + repo + "/releases/latest/download"

	userAgent = "go-agent-updater"
)

// assetName returns the release asset for the current platform, matching the
// goreleaser archive name template (project_name_Os_Arch).
func assetName() (string, error) {
	var osPart string
	switch runtime.GOOS {
	case "linux":
		osPart = "Linux"
	case "darwin":
		osPart = "Darwin"
	case "windows":
		osPart = "Windows"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	var archPart string
	switch runtime.GOARCH {
	case "amd64":
		archPart = "x86_64"
	case "arm64":
		archPart = "arm64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("%s_%s_%s.%s", binName, osPart, archPart, ext), nil
}

// LatestTag fetches the tag of the latest published release.
func LatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiLatest, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("no tag in latest release")
	}
	return rel.TagName, nil
}

// SelfUpdate downloads the latest release archive for this platform and swaps
// the running binary for the new one. It returns once the swap is complete;
// the new version takes effect on the next invocation.
func SelfUpdate(ctx context.Context) error {
	asset, err := assetName()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "go-agent-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, asset)
	if err := download(ctx, dlLatest+"/"+asset, archivePath); err != nil {
		return fmt.Errorf("download %s: %w", asset, err)
	}

	newBin := filepath.Join(tmpDir, binName)
	if strings.HasSuffix(asset, ".zip") {
		err = extractZip(archivePath, binName, newBin)
	} else {
		err = extractTarGz(archivePath, binName, newBin)
	}
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	return replaceExecutable(newBin)
}

func download(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractTarGz(archivePath, wantName, dest string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%q not found in archive", wantName)
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) != wantName {
			continue
		}
		return writeBinary(tr, dest)
	}
}

func extractZip(archivePath, wantName, dest string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zr.Close()

	// Windows binaries carry a .exe suffix inside the archive.
	for _, file := range zr.File {
		base := filepath.Base(file.Name)
		if base != wantName && base != wantName+".exe" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeBinary(rc, dest)
	}
	return fmt.Errorf("%q not found in archive", wantName)
}

func writeBinary(src io.Reader, dest string) error {
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

// replaceExecutable swaps the running binary for newBin. It moves the current
// executable aside first (which succeeds even while running, on both Unix and
// Windows) and rolls back if the swap fails.
func replaceExecutable(newBin string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	dir := filepath.Dir(exe)
	staged := filepath.Join(dir, "."+binName+".new")

	// Copy onto the target filesystem so the final rename is atomic.
	if err := copyFile(newBin, staged); err != nil {
		return fmt.Errorf("stage new binary in %s: %w", dir, err)
	}
	defer os.Remove(staged)

	backup := exe + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(exe, backup); err != nil {
		return fmt.Errorf("move current binary aside: %w", err)
	}

	if err := os.Rename(staged, exe); err != nil {
		_ = os.Rename(backup, exe) // roll back
		return fmt.Errorf("install new binary: %w", err)
	}

	// Best effort; on Windows the running .old can't be deleted until restart.
	_ = os.Remove(backup)
	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// IsNewer reports whether tag is a higher version than current. Both may carry
// a leading "v". A "dev" current is always considered out of date so local
// builds get the notice; unparseable versions fall back to plain inequality.
func IsNewer(tag, current string) bool {
	if current == "" || current == "dev" {
		return true
	}
	t, okT := parseSemver(tag)
	c, okC := parseSemver(current)
	if !okT || !okC {
		return strings.TrimPrefix(tag, "v") != strings.TrimPrefix(current, "v")
	}
	for i := 0; i < 3; i++ {
		if t[i] != c[i] {
			return t[i] > c[i]
		}
	}
	return false
}

func parseSemver(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 { // drop pre-release/build metadata
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

type checkState struct {
	LastCheck time.Time `json:"last_check"`
	LatestTag string    `json:"latest_tag"`
}

func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "go-agent", "update-check.json"), nil
}

func loadCache() checkState {
	var st checkState
	path, err := cachePath()
	if err != nil {
		return st
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return st
	}
	_ = json.Unmarshal(data, &st)
	return st
}

func saveCache(st checkState) {
	path, err := cachePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// NotifyIfOutdated prints a one-line upgrade hint to stderr when a newer
// release exists. The network check runs at most once a day and is cached, so
// most launches add no latency; the hint is printed from the cached tag.
func NotifyIfOutdated(current string) {
	if current == "" || current == "dev" {
		return
	}

	st := loadCache()
	if time.Since(st.LastCheck) > 24*time.Hour {
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		if tag, err := LatestTag(ctx); err == nil {
			st.LatestTag = tag
			st.LastCheck = time.Now()
			saveCache(st)
		}
		cancel()
	}

	if st.LatestTag != "" && IsNewer(st.LatestTag, current) {
		fmt.Fprintf(os.Stderr,
			"\n  A new version of go-agent is available: %s (you have %s)\n  Run \"go-agent update\" to upgrade.\n\n",
			st.LatestTag, current)
	}
}
