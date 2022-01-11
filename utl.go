package mux

import (
	"path"
	"regexp"
	"strings"
)

const (
	namedCaptureGroup    = `(?P<$1>[^\/]+)`
	wildcardCaptureGroup = `(.*)`
)

var patternIdentifiers = [...]string{"*", "{", "}"}

var (
	reCurlyBraces = regexp.MustCompile(`\{([^}]+)\}`)
	reWildcard    = regexp.MustCompile(`\*`)
)

func isPattern(s string) bool {
	for _, c := range patternIdentifiers {
		if strings.Contains(s, c) {
			return true
		}
	}

	return false
}

func patternToRegExp(s string) *regexp.Regexp {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = reWildcard.ReplaceAllString(s, wildcardCaptureGroup)
	s = reCurlyBraces.ReplaceAllString(s, namedCaptureGroup)
	return regexp.MustCompile("^" + s + "$")
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
