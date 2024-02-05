package test_test

import (
	"jinx/server_setup/http_server_setup"
	"testing"
)

func TestFetchResource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		url    string
		expect bool
	}{
		{
			url:    "https://google.com",
			expect: true,
		},
		{
			url:    "https://facebook.com",
			expect: true,
		},
		{
			url:    "/invalid/url",
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			_, err := http_server_setup.FetchResource(test.url)
			if (err != nil) == test.expect {
				t.Errorf("expected %v but got %v", test.expect, err != nil)
			}
		})
	}
}
