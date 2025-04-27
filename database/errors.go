package database

import (
	"net/url"
	"regexp"
)

type NewSourceInstanceError struct {
	Err error
}

func (e *NewSourceInstanceError) Error() string {
	return "could not create new source instance: " + e.sanitizedErr()
}

// make sure we don't expose the database URL in the logs
func (e *NewSourceInstanceError) sanitizedErr() string {
	if e.Err == nil {
		return ""
	}

	// Regular expression to match PostgreSQL connection strings
	pattern := `postgres://[^ \n\t\r]+`
	regExpDB := regexp.MustCompile(pattern)

	// Function to sanitize matched connection strings
	sanitizeURL := func(matchedURL string) string {
		parsedURL, err := url.Parse(matchedURL)
		if err != nil {
			// If parsing fails, return the original matched URL
			return matchedURL
		}

		if parsedURL.User != nil {
			// Replace username and password with '*'
			parsedURL.User = url.UserPassword(`x`, `x`)
		}

		return parsedURL.String()
	}

	// Replace all matched connection strings with their sanitized versions
	sanitizedText := regExpDB.ReplaceAllStringFunc(e.Err.Error(), sanitizeURL)

	return sanitizedText
}
