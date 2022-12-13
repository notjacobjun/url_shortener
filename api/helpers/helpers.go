package helpers

import (
	"os"
	"strings"
)

func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}

	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	// return true if we aren't using a domain that is our current domain (avoiding infinite looping edge case)
	if newURL == os.Getenv("DOMAIN") {
		return false
	}

	return true
}

func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}
