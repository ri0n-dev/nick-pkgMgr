package fetcher

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
)

func GetTarballURL(packageName string) string {
	fmt.Println("[1/4] Fetching metadata for", packageName, "...")
	resp, _ := http.Get("https://registry.npmjs.org/" + packageName)
	defer resp.Body.Close()

	var metadata map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&metadata)

	distTags := metadata["dist-tags"].(map[string]interface{})
	latest := distTags["latest"].(string)

	versions := metadata["versions"].(map[string]interface{})
	latestMeta := versions[latest].(map[string]interface{})
	tarballURL := latestMeta["dist"].(map[string]interface{})["tarball"].(string)

	fmt.Println("[2/4] Metadata fetched successfully")
	return tarballURL
}

func DownloadAndExtract(tarballURL, packageName string) error {
	res, err := http.Get(tarballURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Create a temporary file to store the downloaded content
	tempFile, err := os.CreateTemp("", "npm-download-*.tgz")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Download with progress bar
	bar := progressbar.NewOptions64(
		res.ContentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("[3/4] Downloading %s", packageName)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)

	if _, err := io.Copy(io.MultiWriter(tempFile, bar), res.Body); err != nil {
		return err
	}

	// Reset the file pointer to the beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return err
	}

	// Now extract from the temp file
	gzReader, err := gzip.NewReader(tempFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	fmt.Println("[4/4] Extracting files...")

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join("node_modules", packageName, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			outFile, _ := os.Create(target)
			io.Copy(outFile, tarReader)
			outFile.Close()
		}
	}

	fmt.Println("✅ Success!")
	return nil
}
