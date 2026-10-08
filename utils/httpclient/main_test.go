package httpclient

import (
	"testing"
	"time"
)

// A client without a timeout lets one stalled store hold a search open forever.
func TestClientHasASensibleTimeout(t *testing.T) {
	if Client.Timeout <= 0 || Client.Timeout > 15*time.Second {
		t.Fatalf("expected a timeout between 0 and 15s, got %s", Client.Timeout)
	}
}
