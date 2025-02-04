package util

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const URI_SLASH = "/"
const maxInvalidEscapeRemovalLimit = 10

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

func getInvalidURLEscapeFromError(err error) string {
	errorRegex := regexp.MustCompile("invalid URL escape \"(.*?)\"")
	matches := errorRegex.FindStringSubmatch(err.Error())
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

func hasInvalidURLEscape(parseURL string) (bool, string) {
	_, err := url.Parse(parseURL)
	if err == nil {
		return false, ""
	}

	return isInvalidURLEscapeError(err), getInvalidURLEscapeFromError(err)
}

// UnescapeAllInvalidURLEscapeFromURL - Removes invalid escape characters
// one by one upto max limit.
func UnescapeAllInvalidURLEscapeFromURL(parseURL string) string {
	validURL := parseURL

	removedCount := 10
	for exists, invalidEscape := hasInvalidURLEscape(validURL); exists; {
		removeEscape := strings.ReplaceAll(invalidEscape, "%", "")
		validURL = strings.ReplaceAll(validURL, invalidEscape, removeEscape)
		exists, invalidEscape = hasInvalidURLEscape(validURL)

		// Limit removal of invalid escapes.
		if exists && removedCount == maxInvalidEscapeRemovalLimit {
			log.WithField("url", parseURL).
				Error("Max invalid escape removal reached.")
			return validURL
		}

		removedCount++
	}

	return validURL
}

func isInvalidURLEscapeError(err error) bool {
	return strings.Contains(err.Error(), "invalid URL escape")
}

func ParseURLStable(parseURL string) (*url.URL, error) {
	if !hasProtocol(parseURL) {
		return ParseURLWithoutProtocol(parseURL)
	}

	parsedURL, err := url.Parse(parseURL)
	if err != nil {
		if isInvalidURLEscapeError(err) {
			return url.Parse(UnescapeAllInvalidURLEscapeFromURL(parseURL))
		}

		return nil, err
	}

	return parsedURL, err
}

func TokenizeURI(uri string) []string {
	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(uri, URI_SLASH), URI_SLASH), URI_SLASH)
}

func CleanURI(uri string) string {
	return strings.TrimSuffix(uri, URI_SLASH)
}

func GetURLHostAndPath(pURL *url.URL) string {
	path := GetURLPathWithHash(pURL)

	// removes query params.
	qpSplit := strings.Split(path, "?")
	if len(qpSplit) == 2 {
		path = qpSplit[0]
	}

	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s%s", pURL.Host, path)
}

func GetPathAppendableURLHash(urlHash string) string {
	return strings.Split(urlHash, "?")[0]
}

func GetURLPathWithHash(url *url.URL) string {
	path := url.Path

	if url.Fragment != "" {
		path = path + "#" + url.Fragment
	}

	path = CleanURI(path)
	if path == "" {
		path = fmt.Sprintf("%s/", path)
	}

	return path
}

func GetQueryParamsFromURLFragment(fragment string) map[string]interface{} {
	paramsMap := make(map[string]interface{}, 0)

	if fragment == "" {
		return paramsMap
	}

	ampSplit := strings.Split(fragment, "&")
	for _, keyWithValue := range ampSplit {
		keyValue := strings.Split(keyWithValue, "=")
		if len(keyValue) == 2 && keyValue[0] != "" && keyValue[1] != "" {
			paramsMap[keyValue[0]] = keyValue[1]
		}
	}

	return paramsMap
}
