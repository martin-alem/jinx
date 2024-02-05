package test

import (
	"fmt"
	"jinx/pkg/util/constant"
	"jinx/pkg/util/types"
	"jinx/server_setup/reverse_proxy_server_setup"
	"os"
	"path/filepath"
	"testing"
)

func TestReverseProxySetup(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	routeFile, routeFileErr := os.Create(filepath.Join(tempDir, "route.json"))
	defer func() {
		_ = routeFile.Close()
	}()
	if routeFileErr != nil {
		t.Fatal(routeFileErr)
	}

	_, _ = routeFile.WriteString("{")
	_, _ = routeFile.WriteString(fmt.Sprintf("\"%s\":\"%s\"", "/", "https://google.com"))
	_, _ = routeFile.WriteString("}")

	//Invalid extension
	invalidRouteFile, invalidRouteFileErr := os.Create(filepath.Join(tempDir, "route.txt"))
	defer func() {
		_ = invalidRouteFile.Close()
	}()
	if routeFileErr != nil {
		t.Fatal(invalidRouteFileErr)
	}

	//Invalid content
	invalidRouteContent, invalidRouteContentErr := os.Create(filepath.Join(tempDir, "route2.json"))
	defer func() {
		_ = invalidRouteContent.Close()
	}()
	if routeFileErr != nil {
		t.Fatal(invalidRouteContentErr)
	}

	_, _ = invalidRouteContent.WriteString(fmt.Sprintf("\"%s\":\"%s\"", "/", "https://google.com"))
	_, _ = invalidRouteContent.WriteString("}")

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

	httpsWithRouteTableConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      keyFile.Name(),
		RoutingTable: routeFile.Name(),
	}

	invalidPortConfig := types.ReverseProxyConfig{
		Port:         8080666666699955,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      keyFile.Name(),
		RoutingTable: routeFile.Name(),
	}

	invalidCertFileConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     "/invalid/cert/file",
		KeyFile:      keyFile.Name(),
		RoutingTable: routeFile.Name(),
	}

	invalidKeyFileConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      "/invalid/key/file",
		RoutingTable: routeFile.Name(),
	}

	invalidRouteTableConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      keyFile.Name(),
		RoutingTable: "/invalid/route/table",
	}

	invalidRouteTableFileExtensionConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      keyFile.Name(),
		RoutingTable: invalidRouteFile.Name(),
	}

	invalidRouteTableFileContentConfig := types.ReverseProxyConfig{
		Port:         8080,
		IP:           "127.0.0.1",
		CertFile:     certFile.Name(),
		KeyFile:      keyFile.Name(),
		RoutingTable: invalidRouteContent.Name(),
	}

	tests := []struct {
		name   string
		config types.ReverseProxyConfig
		err    bool
	}{
		{"https setup with valid route table", httpsWithRouteTableConfig, false},
		{"config with invalid port", invalidPortConfig, true},
		{"config with invalid cert", invalidCertFileConfig, true},
		{"config with invalid key file", invalidKeyFileConfig, true},
		{"config with invalid route table", invalidRouteTableConfig, true},
		{"config with invalid route table file extension", invalidRouteTableFileExtensionConfig, true},
		{"config with invalid route table file content", invalidRouteTableFileContentConfig, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := reverse_proxy_server_setup.ReverseProxyServerSetup(test.config, tempDir)
			if (err != nil) != test.err {
				t.Errorf("expected error to %v but go %v", nil, err)
			}

			if test.err == false {
				logRootDir := filepath.Join(tempDir, string(constant.REVERSE_PROXY), constant.LOG_ROOT)
				logRootInfo, logRootStatErr := os.Stat(logRootDir)
				if !logRootInfo.IsDir() || logRootStatErr != nil {
					t.Errorf("expected %s to exist as a directory", logRootDir)
				}
			}
		})
	}
}
