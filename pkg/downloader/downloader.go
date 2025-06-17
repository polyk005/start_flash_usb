package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(url string, savePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Download failed: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
