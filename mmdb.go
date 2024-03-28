package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

const RepoURL = "https://github.com/P3TERX/GeoLite.mmdb"

var Files = []string{"GeoLite2-ASN.mmdb", "GeoLite2-City.mmdb", "GeoLite2-Country.mmdb"}

func GetDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		return path.Join(homeDir, ".mmdb")
	}
	// If os.UserHomeDir fails, create a temporary directory
	tempDir, err := os.MkdirTemp("", "mmdb")
	if err != nil {
		panic(err)
	}
	return tempDir
}

func GetLatest() (tag string, err error) {
	// Create a custom http.Client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Send a HEAD request
	resp, err := client.Head(RepoURL + "/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to send HEAD request: %w", err)
	}
	defer resp.Body.Close()

	// Extract the Location header
	location := resp.Header.Get("location")
	if location == "" {
		return "", fmt.Errorf("Location header is missing")
	}

	// Get the last component of the Location header
	tag = path.Base(location)
	return tag, nil
}

func IsTagDownloaded(dataDir, tag string) bool {
	okFilePath := path.Join(dataDir, tag, ".ok")
	_, err := os.Stat(okFilePath)
	return !os.IsNotExist(err)
}

func DownloadTag(dataDir, tag string) error {
	// Check if the tag has already been downloaded
	if IsTagDownloaded(dataDir, tag) {
		return nil
	}

	// Create directory
	dir := path.Join(dataDir, tag)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download GeoLite2 databases
	for _, file := range Files {
		filePath := path.Join(dir, file)

		// If file exists, continue
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		// Download file
		resp, err := http.Get(RepoURL + "/releases/download/" + tag + "/" + file)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", file, err)
		}
		defer resp.Body.Close()

		out, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file for %s: %w", file, err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write to file for %s: %w", file, err)
		}
	}

	// Create .ok file
	okFilePath := path.Join(dir, ".ok")
	_, err = os.Create(okFilePath)
	if err != nil {
		return fmt.Errorf("failed to create .ok file: %w", err)
	}

	return nil
}

func EnsureLatestDBFiles() ([]string, error) {
	dataDir := GetDataDir()
	latestTag, err := GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %w", err)
	}
	if !IsTagDownloaded(dataDir, latestTag) {
		err = DownloadTag(dataDir, latestTag)
		if err != nil {
			return nil, fmt.Errorf("failed to download latest tag: %w", err)
		}
	}

	// Remove existing symlink if it exists, then create a new one
	symlink := path.Join(dataDir, "latest")
	os.Remove(symlink) // Ignore error, as this will fail if the symlink does not exist
	err = os.Symlink(path.Join(dataDir, latestTag), symlink)
	if err != nil {
		return nil, fmt.Errorf("failed to create symlink to the latest tag: %w", err)
	}

	// Construct paths to the latest database files
	paths := make([]string, len(Files))
	for i, file := range Files {
		paths[i] = path.Join(symlink, file)
	}
	return paths, nil
}
