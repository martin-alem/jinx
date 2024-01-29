package http_server_setup

import (
	"fmt"
	"io"
	"jinx/pkg/util"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func HTTPServerSetup(options map[string]string) {
	port := util.HTTP_PORT

	webRootDir, webRootDirOk := options[util.WEBSITE_ROOT_DIR]
	if !webRootDirOk {
		log.Fatal("website root directory option not specified")
	} else {
		if readable, readableErr := util.IsDirReadable(webRootDir); !readable {
			log.Fatalf("unable to read website directory or does not exit: %s: %v", webRootDir, readableErr)
		}
	}

	ipAddress, ipOk := options[util.IP]
	if !ipOk {
		ipAddress = util.DEFAULT_IP
	}

	certFile, certFileOk := options[util.CERT_FILE]
	if certFileOk && certFile != "" {
		if readable, readableErr := util.IsDirReadable(certFile); !readable {
			log.Fatalf("%s: %v", certFile, readableErr)
		}
	}

	keyFile, keyFileOk := options[util.KEY_FILE]
	if keyFileOk && keyFile != "" {
		if readable, readableErr := util.IsDirReadable(keyFile); !readable {
			log.Fatalf("%s: %v", keyFile, readableErr)
		}
	}

	if certFileOk && certFile != "" && keyFileOk && keyFile != "" {
		port = util.HTTPS_PORT
	}

	defaultWebsiteRoot := filepath.Join(util.BASE, util.DEFAULT_WEBSITE_ROOT)
	if mkdirErr := os.Mkdir(defaultWebsiteRoot, 0755); mkdirErr != nil {
		log.Fatalf("unable to create default website root: %v", mkdirErr)
	}

	imagesDir := filepath.Join(defaultWebsiteRoot, util.IMAGE_DIR)
	if mkdirErr := os.Mkdir(imagesDir, 0755); mkdirErr != nil {
		log.Fatalf("unable to initialize default website image dir: %v", mkdirErr)
	}

	resources := map[string]string{
		util.JINX_INDEX_URL: util.JINX_INDEX_FILE,
		util.JINX_404_URL:   util.JINX_404_FILE,
		util.JINX_CSS_URL:   util.JINX_CSS_FILE,
		util.JINX_ICO_URL:   util.JINX_ICO_FLE,
		util.JINX_SVG_URL:   util.JINX_SVG_FILE,
	}

	var wg sync.WaitGroup
	wg.Add(len(resources)) // Add count to WaitGroup before starting goroutines

	resourceChan := make(chan util.JinxResourceResponse, len(resources))

	for url, file := range resources {
		go func(resourceURL string, fileName string) {
			defer wg.Done() // Ensure wg.Done() is called when goroutine finishes
			res, err := http.Get(resourceURL)
			if err != nil {
				log.Printf("unable to fetch resource from URL %s: %v", resourceURL, err)
				return
			}
			resourceChan <- util.JinxResourceResponse{Res: res, Filename: fileName}
		}(url, file)
	}

	// Close resourceChan after all goroutines have finished
	go func() {
		wg.Wait()
		close(resourceChan)
	}()

	// Process received resources
	for data := range resourceChan {
		HandleResourceResponse(defaultWebsiteRoot, imagesDir, &data)
		_ = data.Res.Body.Close()
	}

	configuration := map[string]any{
		"ip":               ipAddress,
		"port":             port,
		"cert-file":        certFile,
		"key-file":         keyFile,
		"website-root-dir": webRootDir,
	}

	configPath := filepath.Join(util.BASE, util.CONFIG_FILE)
	configFileHandle, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, 0644)
	defer func() {
		_ = configFileHandle.Close()
	}()

	if err != nil {
		_ = os.RemoveAll(util.BASE)
		log.Fatalf("unable to create config file for http server: %v", err)
	}

	if _, writeErr := configFileHandle.WriteString(fmt.Sprintf("%s", configuration)); writeErr != nil {
		_ = os.RemoveAll(util.BASE)
		log.Fatalf("unable to write configuration to file: %v", err)
	}

}

func HandleResourceResponse(websiteRoot string, imageDir string, resource *util.JinxResourceResponse) {

	fileContent, err := io.ReadAll(resource.Res.Body)
	if err != nil {
		log.Fatalf("unable to read response for: %v", resource.Filename)
	}

	filePath := filepath.Join(websiteRoot, resource.Filename)

	if resource.Filename == util.JINX_ICO_FLE || resource.Filename == util.JINX_SVG_FILE {
		filePath = filepath.Join(imageDir, resource.Filename)
	}

	fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	defer func() {
		_ = fileHandle.Close()
	}()
	if err != nil {
		log.Fatalf("unable to open file %s: %v", filePath, err)
	}

	if _, writeErr := fileHandle.Write(fileContent); writeErr != nil {
		log.Fatalf("error writing to %s: %v", filePath, writeErr)
	}
}
