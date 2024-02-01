package test

import (
	"jinx/pkg/util/helper"
	"testing"
)

func TestSingleJoiningSlash(t *testing.T) {

	tests := []struct {
		name   string
		base   string
		path   string
		result string
	}{
		{name: "BaseAndPathHasSlashes", base: "http://mysite.com/", path: "/pages/about", result: "http://mysite.com/pages/about"},
		{name: "BaseNoTrailingSlashPathHasLeadingSlash", base: "http://mysite.com", path: "/pages/about", result: "http://mysite.com/pages/about"},
		{name: "BaseHasTrailingSlashPathNoLeadingSlash", base: "http://mysite.com/", path: "pages/about", result: "http://mysite.com/pages/about"},
		{name: "BaseAndPathNoSlashes", base: "http://mysite.com", path: "pages/about", result: "http://mysite.com/pages/about"},
		{name: "BaseAndPathEmpty", base: "http://mysite.com", path: "", result: "http://mysite.com/"},
		{name: "BaseEmptyPathHasSlash", base: "", path: "/pages/about", result: "/pages/about"},
		{name: "BaseEmptyPathNoSlash", base: "", path: "pages/about", result: "/pages/about"},
		{name: "BaseWithPathQueryParams", base: "http://mysite.com/", path: "/pages/about?section=team", result: "http://mysite.com/pages/about?section=team"},
		{name: "BaseWithPathAndFragment", base: "http://mysite.com/", path: "/pages/about#team", result: "http://mysite.com/pages/about#team"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := helper.SingleJoiningSlash(test.base, test.path); result != test.result {
				t.Errorf("expected %s but got %s", test.result, result)
			}

		})
	}
}
