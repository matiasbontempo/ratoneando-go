package controllers

import (
	"strings"

	"ratoneando/config"
)

// isAllowedReferer lets requests through in development, and in release only when they come
// from the web app.
func isAllowedReferer(referer string) bool {
	if config.ENV != "release" {
		return true
	}
	return referer != "" && strings.Contains(referer, config.WEB_URL)
}
