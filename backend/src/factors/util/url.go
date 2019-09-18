package util

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const URI_SLASH = "/"

var regex_URI_LAST_SLASH_SUFFIX = regexp.MustCompile("(/.+)(/[^/]+)/?$")

// PopURIBySlash pop last part of uri by uri_slash using regex.
// Input: uri=/u1/u2/u3
// Returns: /u1/u2, /u3 (afterPopURI, poppedURIPart)
func PopURIBySlash(uri string) (string, string) {
	if len(uri) == 0 {
		return "", ""
	}

	// Returns [originalURI, afterPoppedURI, poppedURIPart]
	poppedGroup := regex_URI_LAST_SLASH_SUFFIX.FindStringSubmatch(uri)

	// Handling last part of uri. /u1 -> "", /u1
	if len(poppedGroup) < 3 {
		return "", uri
	}

	return poppedGroup[1], poppedGroup[2]
}

func hasProtocol(purl string) bool {
	return len(strings.Split(purl, "://")) > 1
}

func ParseURLWithoutProtocol(parseURL string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("dummy://%s", parseURL))
}

func ParseURLStable(parseURL string) (*url.URL, error) {
	if !hasProtocol(parseURL) {
		return ParseURLWithoutProtocol(parseURL)
	}
	return url.Parse(parseURL)
}

func TokenizeURI(uri string) []string {
	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(uri, URI_SLASH), URI_SLASH), URI_SLASH)
}

func CleanURI(uri string) string {
	return strings.TrimSuffix(uri, URI_SLASH)
}

func GetURLHostAndPath(parseURL string) (string, error) {
	cURL := strings.TrimSpace(parseURL)

	if cURL == "" {
		return "", errors.New("parsing failed empty url")
	}

	pURL, err := ParseURLStable(cURL)
	if err != nil {
		return "", err
	}

	// adds / as suffix for root.
	path := GetURLPathWithHash(pURL)
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s%s", pURL.Host, path), nil
}

func GetPathAppendableURLHash(urlHash string) string {
	return strings.Split(urlHash, "?")[0]
}

func GetURLPathWithHash(url *url.URL) string {
	path := url.Path
	if url.Fragment != "" {
		// URL fragment removes #. added # back.
		path = path + "#" + url.Fragment
	}

	return CleanURI(path)
}
