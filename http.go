package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/schollz/progressbar/v3"
)

const TempDir = "temp/"

func DownloadFile(url string, targetPath string) (string, error) {
	var filepath string

	if strings.HasSuffix(targetPath, "/") {
		_, filename := path.Split(url)
		filepath = path.Join(targetPath, filename)
	} else {
		filepath = targetPath
		targetPath, _ = path.Split(filepath)
	}

	if !IsDirExists(targetPath) {
		err := os.MkdirAll(targetPath, os.ModeDir)
		if err != nil {
			return "", err
		}
	}

	if IsFileExists(filepath) {
		Warn.Printf("File already exists: %s", filepath)
		return filepath, os.ErrExist
	}

	Info.Println("Start downloading...")
	Info.Println("\t", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Referer", Referer)
	req.Header.Set("Origin", Origin)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer CloseQuietly(f)

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	defer CloseQuietly(resp.Body)
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return "", err
	}

	Info.Println("Download finished")

	return filepath, nil
}

func HttpOpen(url string, cachePath string) (file *os.File, err error) {
	filepath, err := DownloadFile(url, cachePath)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	file, err = os.Open(filepath)
	if err != nil {
		return nil, err
	}
	// defer CloseQuietly(file)

	return file, nil
}
