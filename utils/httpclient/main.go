package httpclient

import (
	"net/http"
	"time"
)

// Client is used by every scraper. http.DefaultClient has no timeout, so a store that stops
// answering would hold a whole search open indefinitely.
var Client = &http.Client{Timeout: 8 * time.Second}
