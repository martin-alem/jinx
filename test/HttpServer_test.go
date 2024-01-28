package test

import (
	"fmt"
	"io"
	"jinx/internal/jinx_http"
	"jinx/pkg/util"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const defaultIndexFileContent = "<html><head><title>Index</title></head><body>Hello World! Default Index Page</body></html>"
const customIndexFileContent = "<html><head><title>Index</title></head><body>Hello World! Custom Index Page</body></html>"
const defaultNotFoundContent = "<html><head><title>404</title></head><body>Default Page Not Found</body></html>"
const customNotFoundContent = "<html><head><title>404</title></head><body>Custom Page Not Found</body></html>"
const aboutFileContents = "<html><head><title>About</title></head><body>This is about my website</body></html>"

func CompleteServerSetup(t *testing.T) (handler http.Handler, dir string) {

	tempDir, err := os.MkdirTemp("", "jinx_root")
	if err != nil {
		t.Fatal(err)
	}

	serverRoot := filepath.Join(tempDir, "jinx")
	if err := os.Mkdir(serverRoot, 0755); err != nil {
		t.Fatal(err)
	}

	defaultWebRoot := filepath.Join(serverRoot, "www")
	if err := os.Mkdir(defaultWebRoot, 0755); err != nil {
		t.Fatal(err)
	}

	indexFile := filepath.Join(defaultWebRoot, "index.html")
	indexFileHandle, err := os.OpenFile(indexFile, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = indexFileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := indexFileHandle.WriteString(defaultIndexFileContent); err != nil {
		t.Fatal(err)
	}

	notFoundFile := filepath.Join(defaultWebRoot, "404.html")
	notFoundFileHandle, err := os.OpenFile(notFoundFile, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = notFoundFileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notFoundFileHandle.WriteString(defaultNotFoundContent); err != nil {
		t.Fatal(err)
	}

	webRoot := filepath.Join(tempDir, "websites")
	if err := os.Mkdir(webRoot, 0755); err != nil {
		t.Fatal(err)
	}

	websiteRoot := filepath.Join(webRoot, "mysite.com")
	if err := os.Mkdir(websiteRoot, 0755); err != nil {
		t.Fatal(err)
	}

	pagesDir := filepath.Join(websiteRoot, "pages")
	if err := os.Mkdir(pagesDir, 0755); err != nil {
		t.Fatal(err)
	}

	indexFile2 := filepath.Join(websiteRoot, "index.html")
	indexFileHandle2, err := os.OpenFile(indexFile2, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = indexFileHandle2.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := indexFileHandle2.WriteString(customIndexFileContent); err != nil {
		t.Fatal(err)
	}

	notFoundFile2 := filepath.Join(websiteRoot, "404.html")
	notFoundFileHandle2, err := os.OpenFile(notFoundFile2, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = notFoundFileHandle2.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notFoundFileHandle2.WriteString(customNotFoundContent); err != nil {
		t.Fatal(err)
	}

	aboutFile := filepath.Join(pagesDir, "about.html")
	aboutFileHandle, err := os.OpenFile(aboutFile, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = aboutFileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aboutFileHandle.WriteString(aboutFileContents); err != nil {
		t.Fatal(err)
	}

	logRoot := filepath.Join(tempDir, "logs")
	if err := os.Mkdir(logRoot, 0755); err != nil {
		t.Fatal(err)
	}

	config := util.JinxHttpServerConfig{
		IP:          "127.0.0.1",
		Port:        8080,
		LogRoot:     logRoot,
		WebsiteRoot: webRoot,
	}

	jinx := jinx_http.NewJinxHttpServer(config, serverRoot)

	return jinx, tempDir
}

// Installed Server Software Correctly, Server Root Dir Exist But User Has No Websites
func InCompleteServerSetup(t *testing.T) (handler http.Handler, dir string) {

	tempDir, err := os.MkdirTemp("", "jinx_root")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	serverRoot := filepath.Join(tempDir, "jinx")
	if err := os.Mkdir(serverRoot, 0755); err != nil {
		t.Fatal(err)
	}

	defaultWebRoot := filepath.Join(serverRoot, "www")
	if err := os.Mkdir(defaultWebRoot, 0755); err != nil {
		t.Fatal(err)
	}

	indexFile := filepath.Join(defaultWebRoot, "index.html")
	indexFileHandle, err := os.OpenFile(indexFile, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = indexFileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := indexFileHandle.WriteString(defaultIndexFileContent); err != nil {
		t.Fatal(err)
	}

	notFoundFile := filepath.Join(defaultWebRoot, "404.html")
	notFoundFileHandle, err := os.OpenFile(notFoundFile, os.O_RDWR|os.O_CREATE, 0644)
	defer func() {
		_ = notFoundFileHandle.Close()
	}()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notFoundFileHandle.WriteString(defaultNotFoundContent); err != nil {
		t.Fatal(err)
	}

	logRoot := filepath.Join(tempDir, "logs")
	if err := os.Mkdir(logRoot, 0755); err != nil {
		t.Fatal(err)
	}

	config := util.JinxHttpServerConfig{
		IP:          "127.0.0.1",
		Port:        8080,
		LogRoot:     logRoot,
		WebsiteRoot: "/random/web/root",
	}

	jinx := jinx_http.NewJinxHttpServer(config, serverRoot)

	return jinx, tempDir

}

func TestJinxHttpServerWithCompleteSetup(t *testing.T) {

	jinx, dir := CompleteServerSetup(t)
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	server := httptest.NewServer(jinx)
	defer server.Close()

	tests := []struct {
		host                string
		path                string
		expectedBody        string
		expectedStatusCode  int
		expectedContentType string
	}{
		{
			"127.0.0.1",
			"/",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"127.0.0.1",
			"",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"127.0.0.1",
			"/man.html",
			defaultNotFoundContent,
			404,
			"text/html",
		},
		{
			"mysite.com",
			"/",
			customIndexFileContent,
			200,
			"text/html",
		},
		{
			"mysite.com",
			"",
			customIndexFileContent,
			200,
			"text/html",
		},
		{
			"mysite.com",
			"/pages/about.html",
			aboutFileContents,
			200,
			"text/html",
		},
		{
			"mysite.com",
			"/pages/about.htm",
			customNotFoundContent,
			404,
			"text/html",
		},
		{
			"mysite.com",
			"/about.html",
			customNotFoundContent,
			404,
			"text/html",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s:%s", test.host, test.path), func(t *testing.T) {
			// Create a new request to the specified path on the httptest.Server
			req, err := http.NewRequest(http.MethodGet, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Host = test.host

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = res.Body.Close()
			}()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			resContent := string(body)
			statusCode := res.StatusCode
			contentType := res.Header.Get("Content-Type")

			if resContent != test.expectedBody {
				t.Errorf("expected %s got %s", test.expectedBody, resContent)
			}

			if statusCode != test.expectedStatusCode {
				t.Errorf("expected %d got %d", test.expectedStatusCode, statusCode)
			}

			if !strings.Contains(contentType, test.expectedContentType) {
				t.Errorf("expected %s got %s", test.expectedContentType, contentType)
			}
		})
	}

}

func TestJinxHttpServerWithInCompleteSetup(t *testing.T) {
	jinx, dir := InCompleteServerSetup(t)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	server := httptest.NewServer(jinx)
	defer server.Close()

	tests := []struct {
		host                string
		path                string
		expectedBody        string
		expectedStatusCode  int
		expectedContentType string
	}{
		{
			"127.0.0.1",
			"/",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"127.0.0.1",
			"",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"127.0.0.1",
			"/man.html",
			defaultNotFoundContent,
			404,
			"text/html",
		},
		{
			"mysite.com",
			"/",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"mysite.com",
			"",
			defaultIndexFileContent,
			200,
			"text/html",
		},
		{
			"mysite.com",
			"/pages/about.html",
			defaultNotFoundContent,
			404,
			"text/html",
		},
		{
			"mysite.com",
			"/pages/about.htm",
			defaultNotFoundContent,
			404,
			"text/html",
		},
		{
			"mysite.com",
			"/about.html",
			defaultNotFoundContent,
			404,
			"text/html",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s:%s", test.host, test.path), func(t *testing.T) {
			// Create a new request to the specified path on the httptest.Server
			req, err := http.NewRequest(http.MethodGet, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Host = test.host

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = res.Body.Close()
			}()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			resContent := string(body)
			statusCode := res.StatusCode
			contentType := res.Header.Get("Content-Type")

			if resContent != test.expectedBody {
				t.Errorf("expected %s got %s", test.expectedBody, resContent)
			}

			if statusCode != test.expectedStatusCode {
				t.Errorf("expected %d got %d", test.expectedStatusCode, statusCode)
			}

			if !strings.Contains(contentType, test.expectedContentType) {
				t.Errorf("expected %s got %s", test.expectedContentType, contentType)
			}
		})
	}

}
