package history

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

type testStore struct {
	*Store
	clock time.Time
}

func newTestStore(t *testing.T) *testStore {
	t.Helper()

	store, err := Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	ts := &testStore{Store: store, clock: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	store.now = func() time.Time { return ts.clock }
	return ts
}

func (ts *testStore) advance(d time.Duration) { ts.clock = ts.clock.Add(d) }

func (ts *testStore) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := ts.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (ts *testStore) observations(t *testing.T) int {
	return ts.count(t, "SELECT COUNT(*) FROM observation")
}

func item(price float64) Item {
	return Item{Source: "disco", ID: "1", Name: "Leche 1 L", Unit: "LT", UnitFactor: 1, Price: price}
}

func record(t *testing.T, ts *testStore, items ...Item) Result {
	t.Helper()
	result, err := ts.Record(items)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestFirstReadingIsStored(t *testing.T) {
	ts := newTestStore(t)

	result := record(t, ts, item(1000))

	if result.Inserted != 1 || ts.observations(t) != 1 {
		t.Fatalf("expected 1 stored observation, got result %+v", result)
	}
	if ts.count(t, "SELECT COUNT(*) FROM sku") != 1 {
		t.Fatal("expected the sku to be created")
	}
}

func TestUnchangedPriceIsNotStoredAgain(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	ts.advance(time.Hour)
	result := record(t, ts, item(1000))

	if result.Unchanged != 1 || ts.observations(t) != 1 {
		t.Fatalf("expected the unchanged price to be skipped, got %+v", result)
	}
}

func TestChangedPriceIsStored(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	ts.advance(time.Hour)
	record(t, ts, item(1100))

	if ts.observations(t) != 2 {
		t.Fatalf("expected 2 observations, got %d", ts.observations(t))
	}
}

func TestListPriceChangeIsStored(t *testing.T) {
	ts := newTestStore(t)
	first := item(1000)
	first.ListPrice = 1200
	record(t, ts, first)

	ts.advance(time.Hour)
	second := item(1000)
	second.ListPrice = 1500
	record(t, ts, second)

	if ts.observations(t) != 2 {
		t.Fatalf("expected the list price change to be stored, got %d", ts.observations(t))
	}
}

func TestHeartbeatStoresAnUnchangedPriceAfterADay(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	ts.advance(25 * time.Hour)
	record(t, ts, item(1000))

	if ts.observations(t) != 2 {
		t.Fatalf("expected a heartbeat observation, got %d", ts.observations(t))
	}
}

func TestInvalidReadingsAreSkipped(t *testing.T) {
	ts := newTestStore(t)

	noID := item(1000)
	noID.ID = ""
	noSource := item(1000)
	noSource.Source = ""

	result := record(t, ts, noID, noSource, item(0), item(-5))

	if result.Skipped != 4 || ts.observations(t) != 0 {
		t.Fatalf("expected 4 skipped readings and nothing stored, got %+v", result)
	}
}

func TestSuddenJumpIsHeldBackUntilItRepeats(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	// A glitchy reading: 10x the previous price.
	ts.advance(time.Hour)
	if result := record(t, ts, item(10000)); result.Suspicious != 1 {
		t.Fatalf("expected the first odd reading to be held back, got %+v", result)
	}
	// The price goes back to normal, which clears the suspicion.
	ts.advance(time.Hour)
	record(t, ts, item(1000))
	ts.advance(time.Hour)
	if result := record(t, ts, item(10000)); result.Suspicious != 1 {
		t.Fatalf("expected the count to restart after a normal reading, got %+v", result)
	}
	if ts.observations(t) != 1 {
		t.Fatalf("expected the glitch to never be stored, got %d observations", ts.observations(t))
	}
}

func TestSustainedJumpIsEventuallyAccepted(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	for i := 0; i < suspectConfirmations-1; i++ {
		ts.advance(time.Hour)
		record(t, ts, item(5000))
	}
	if ts.observations(t) != 1 {
		t.Fatal("the jump should still be held back before it is confirmed")
	}

	ts.advance(time.Hour)
	result := record(t, ts, item(5000))

	if result.Inserted != 1 || ts.observations(t) != 2 {
		t.Fatalf("expected the confirmed price to be stored, got %+v", result)
	}
}

func TestSkuDetailsAreUpdated(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	renamed := item(1000)
	renamed.Name = "Leche Entera 1 L"
	ts.advance(time.Hour)
	record(t, ts, renamed)

	var name string
	var firstSeen, lastSeen int64
	err := ts.db.QueryRow("SELECT name, first_seen, last_seen FROM sku").Scan(&name, &firstSeen, &lastSeen)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Leche Entera 1 L" || lastSeen-firstSeen != 3600 {
		t.Fatalf("unexpected sku row: %s %d %d", name, firstSeen, lastSeen)
	}
}

func TestSameProductInDifferentStoresIsIndependent(t *testing.T) {
	ts := newTestStore(t)
	other := item(1000)
	other.Source = "vea"

	record(t, ts, item(1000), other)

	if ts.count(t, "SELECT COUNT(*) FROM sku") != 2 || ts.observations(t) != 2 {
		t.Fatal("expected one sku and one observation per store")
	}
}

func TestBackupProducesAReadableCopy(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	if err := ts.Backup(); err != nil {
		t.Fatal(err)
	}
	// Running it twice must work: the previous copy is replaced.
	if err := ts.Backup(); err != nil {
		t.Fatal(err)
	}

	copyDB, err := sql.Open("sqlite", ts.path+".backup")
	if err != nil {
		t.Fatal(err)
	}
	defer copyDB.Close()

	var n int
	if err := copyDB.QueryRow("SELECT COUNT(*) FROM observation").Scan(&n); err != nil || n != 1 {
		t.Fatalf("expected the backup to contain the observation, got %d (%v)", n, err)
	}
}

func TestPurgeSourcesDeletesOnlyTheNamedSources(t *testing.T) {
	ts := newTestStore(t)
	other := item(1000)
	other.Source = "vea"
	keep := item(1000)
	keep.Source = "jumbo"
	record(t, ts, item(1000), other, keep)
	ts.advance(time.Hour)
	record(t, ts, item(1100), other)

	observations, skus, err := ts.PurgeSources([]string{"disco", "vea"})
	if err != nil {
		t.Fatal(err)
	}

	if skus != 2 || observations != 3 {
		t.Fatalf("expected 2 products and 3 readings deleted, got %d and %d", skus, observations)
	}
	if ts.count(t, "SELECT COUNT(*) FROM sku WHERE source = 'jumbo'") != 1 ||
		ts.count(t, "SELECT COUNT(*) FROM observation WHERE source = 'jumbo'") != 1 {
		t.Fatal("the other source must be left alone")
	}
	if ts.count(t, "SELECT COUNT(*) FROM observation WHERE source IN ('disco','vea')") != 0 {
		t.Fatal("expected nothing left for the purged sources")
	}
}

func TestPurgeWithNoMatchingSourceDoesNothing(t *testing.T) {
	ts := newTestStore(t)
	record(t, ts, item(1000))

	observations, skus, err := ts.PurgeSources([]string{"nowhere"})

	if err != nil || observations != 0 || skus != 0 || ts.observations(t) != 1 {
		t.Fatalf("unexpected purge result: %d %d %v", observations, skus, err)
	}
}
