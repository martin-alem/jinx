package test_test

import (
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"jinx/server_setup/http_server_setup"
	"os"
	"path/filepath"
	"testing"
)

func TestHTTPServerSetup(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	//Create Key File in root
	keyFile, keyFileErr := os.Create(filepath.Join(tempDir, "private.key"))
	defer func() {
		_ = keyFile.Close()
	}()
	if keyFileErr != nil {
		t.Fatal(keyFileErr)
	}

	//Create CertFile in root
	certFile, certFileErr := os.Create(filepath.Join(tempDir, "certificate.crt"))
	defer func() {
		_ = certFile.Close()
	}()
	if certFileErr != nil {
		t.Fatal(certFileErr)
	}

	//Create website directory
	webDirErr := os.Mkdir(filepath.Join(tempDir, "websites"), 0755)
	if webDirErr != nil {
		t.Fatal(webDirErr)
	}

	httpsWithWebDirConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       certFile.Name(),
		KeyFile:        keyFile.Name(),
		WebsiteRootDir: filepath.Join(tempDir, "websites"),
	}

	invalidPortConfig := types.HttpServerConfig{
		Port:           8080699999999999,
		IP:             "127.0.0.1",
		CertFile:       certFile.Name(),
		KeyFile:        keyFile.Name(),
		WebsiteRootDir: filepath.Join(tempDir, "websites"),
	}

	invalidCertFileConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       "/invalid/path",
		KeyFile:        keyFile.Name(),
		WebsiteRootDir: filepath.Join(tempDir, "websites"),
	}

	invalidKeyFileConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       certFile.Name(),
		KeyFile:        "/invalid/path",
		WebsiteRootDir: filepath.Join(tempDir, "websites"),
	}

	invalidWebDirConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       certFile.Name(),
		KeyFile:        keyFile.Name(),
		WebsiteRootDir: "/invalid/path",
	}

	httpsNoWebDirConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       certFile.Name(),
		KeyFile:        keyFile.Name(),
		WebsiteRootDir: "",
	}

	httpNoWebDirConfig := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       "",
		KeyFile:        "",
		WebsiteRootDir: "",
	}

	httpWithValidWebDir := types.HttpServerConfig{
		Port:           8080,
		IP:             "127.0.0.1",
		CertFile:       "",
		KeyFile:        "",
		WebsiteRootDir: filepath.Join(tempDir, "websites"),
	}

	tests := []struct {
		name       string
		config     types.HttpServerConfig
		errorCode  int
		checkSetup bool
	}{
		{name: "https config with valid web dir", config: httpsWithWebDirConfig, errorCode: 0, checkSetup: true},
		{name: "config with invalid port", config: invalidPortConfig, errorCode: constant.INVALID_PORT, checkSetup: false},
		{name: "config with invalid cert file path", config: invalidCertFileConfig, errorCode: constant.INVALID_CERT_PATH, checkSetup: false},
		{name: "config with invalid key file path", config: invalidKeyFileConfig, errorCode: constant.INVALID_KEY_PATH, checkSetup: false},
		{name: "config with invalid web dir path", config: invalidWebDirConfig, errorCode: constant.INVALID_WEBSITE_DIR, checkSetup: false},
		{name: "https config no web dir", config: httpsNoWebDirConfig, errorCode: 0, checkSetup: true},
		{name: "http config no web dir", config: httpNoWebDirConfig, errorCode: 0, checkSetup: true},
		{name: "http with valid web dir", config: httpWithValidWebDir, errorCode: 0, checkSetup: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := http_server_setup.HTTPServerSetup(test.config, tempDir)

			if test.errorCode == 0 && err != nil {
				t.Errorf("expected error to %v but go %v", nil, err)
			}

			if test.errorCode != 0 && err.ErrorCode != test.errorCode {
				t.Errorf("expected %v but got %v", test.errorCode, err.ErrorCode)
			}

			if test.checkSetup == true {
				logRootDir := filepath.Join(tempDir, string(constant.HTTP_SERVER), constant.LOG_ROOT)
				logRootInfo, logRootStatErr := os.Stat(logRootDir)
				if !logRootInfo.IsDir() || logRootStatErr != nil {
					t.Errorf("expected %s to exist as a directory", logRootDir)
				}

				//Test default website directory is created
				defaultWebDir := filepath.Join(tempDir, string(constant.HTTP_SERVER), constant.DEFAULT_WEBSITE_ROOT)
				defaultWebDirInfo, defaultWebDirErr := os.Stat(defaultWebDir)
				if !defaultWebDirInfo.IsDir() || defaultWebDirErr != nil {
					t.Errorf("expected %s to exist as a directory", defaultWebDir)
				}

				//Test images dir exist in default website dir
				imageDir := filepath.Join(defaultWebDir, constant.IMAGE_DIR)
				imageDirInfo, imageDirErr := os.Stat(imageDir)
				if !imageDirInfo.IsDir() || imageDirErr != nil {
					t.Errorf("expected %s to exist as a directory", imageDir)
				}

				//Test index.html, 404.html, stylesheet.css exist in website
				webFiles := []string{
					filepath.Join(defaultWebDir, constant.JINX_INDEX_FILE),
					filepath.Join(defaultWebDir, constant.JINX_404_FILE),
					filepath.Join(defaultWebDir, constant.JINX_CSS_FILE),
				}

				for _, file := range webFiles {
					if fileInfo, fileInfoErr := os.Stat(file); fileInfo.IsDir() || fileInfoErr != nil {
						t.Errorf("expected %s to exist as a file", file)
					}
				}

				//Test jinx.ico and jinx.svg exist in images
				imageFiles := []string{
					filepath.Join(imageDir, "jinx.ico"),
					filepath.Join(imageDir, "jinx.svg"),
				}
				for _, file := range imageFiles {
					if imageFileInfo, imageFileErr := os.Stat(file); imageFileInfo.IsDir() || imageFileErr != nil {
						t.Errorf("expected %s to exist as a file", file)
					}
				}
			}
		})
	}

}
