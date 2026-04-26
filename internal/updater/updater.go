package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repo    = "yusupkhemraev/argus"
	apiURL  = "https://api.github.com/repos/" + repo + "/releases/latest"
	timeout = 30 * time.Second
)

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func assetName() string {
	return fmt.Sprintf("argus-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func fetchLatest() (*Release, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}
	return &release, nil
}

func Update(currentVersion string) (string, error) {
	release, err := fetchLatest()
	if err != nil {
		return "", err
	}

	latest := release.TagName
	cur := strings.TrimPrefix(currentVersion, "v")
	lat := strings.TrimPrefix(latest, "v")

	if cur == lat {
		return latest, fmt.Errorf("already on latest version %s", latest)
	}

	name := assetName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == name {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return latest, fmt.Errorf("no asset found for %s", name)
	}

	exe, err := os.Executable()
	if err != nil {
		return latest, fmt.Errorf("find executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return latest, fmt.Errorf("resolve symlink: %w", err)
	}

	tmp := exe + ".new"
	if err := download(downloadURL, tmp); err != nil {
		return latest, err
	}

	if err := os.Chmod(tmp, 0755); err != nil {
		os.Remove(tmp)
		return latest, fmt.Errorf("chmod: %w", err)
	}

	// atomic replace — rename works even on running binary on Linux
	if err := os.Rename(tmp, exe); err != nil {
		os.Remove(tmp)
		return latest, fmt.Errorf("replace binary: %w", err)
	}

	return latest, nil
}

func download(url, dst string) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: status %d", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(dst)
		return fmt.Errorf("write binary: %w", err)
	}
	return nil
}
