package utils

import (
	"downloader/internal/config"
	"downloader/pkg/LOG"
	"downloader/pkg/utils/barutils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/vbauerster/mpb/v8"
)

var DefaultClient *http.Client

func init() {
	DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

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
		if err := os.MkdirAll(targetPath, os.ModeDir); err != nil {
			return "", err
		}
	}

	if IsFileExists(filepath) {
		LOG.Warn.Printf("File already exists: %s", filepath)
		return filepath, os.ErrExist
	}

	LOG.Info.Println("Start downloading...")
	LOG.Info.Println("\t", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("Origin", config.Origin)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad http status: %s %s", resp.Proto, resp.Status)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer CloseQuietly(f)

	bar := barutils.NewProgressBarBytes(
		resp.ContentLength,
		"Downloading",
	)

	defer CloseQuietly(resp.Body)
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return "", err
	}

	LOG.Info.Println("Download finished")

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

func MultiDownload(urls []string, dir string, numThreads int) error {
	var wg sync.WaitGroup
	var sem = make(chan struct{}, numThreads)
	var ec = make(chan error, len(urls))

	p := barutils.NewProgress(&wg, config.BarWidth, config.BarRefreshRate)

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(url string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			if err := download(url, dir, p); err != nil {
				ec <- err
			}
		}(url)
	}

	p.Wait()
	close(ec)

	select {
	case err := <-ec:
		return err
	default:
		return nil
	}
}

func download(url, dir string, p *mpb.Progress) (err error) {
	filename := url[strings.LastIndex(url, "/")+1:]
	filePath := path.Join(dir, filename)
	if IsFileExists(filePath) {
		LOG.Warn.Printf("Using cache: %s", filePath)
		return
	}
	if err = os.MkdirAll(path.Dir(filePath), os.ModePerm); err != nil {
		return
	}

	var req *http.Request
	var resp *http.Response
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Referer", config.Referer)
	req.Header.Set("Origin", config.Origin)

	if resp, err = DefaultClient.Do(req); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer CloseQuietly(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad http status: %s %s", resp.Proto, resp.Status)
	}

	var contentLength int
	if contentLength, err = strconv.Atoi(resp.Header.Get("Content-Length")); err != nil {
		return fmt.Errorf("bad `Content-Length`: \"%s\"", resp.Header.Get("Content-Length"))
	}

	bar := barutils.NewBar(p, int64(contentLength), fmt.Sprintf("Downloading %s: ", filename))

	proxyReader := bar.ProxyReader(resp.Body)
	defer CloseQuietly(proxyReader)

	var output *os.File
	if output, err = os.Create(filePath); err != nil {
		return fmt.Errorf("create output failed: %w", err)
	}
	if _, err = io.Copy(output, proxyReader); err != nil {
		_ = os.Remove(filename)
		return
	}

	return
}
