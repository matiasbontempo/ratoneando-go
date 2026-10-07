package history

import (
	"testing"
	"time"

	"ratoneando/config"
	"ratoneando/products"
)

func useDefaults(t *testing.T) {
	t.Helper()
	config.HISTORY_WINDOW_DAYS = 60
	config.HISTORY_MIN_POINTS = 5
	config.HISTORY_MIN_SPAN_DAYS = 14
}

// seed stores observations for one product. daysAgo is relative to the test clock, and the
// first entry decides first_seen.
func seed(t *testing.T, ts *testStore, source, id string, points map[int]float64) {
	t.Helper()

	oldest := 0
	for daysAgo := range points {
		if daysAgo > oldest {
			oldest = daysAgo
		}
	}
	first := ts.clock.Add(-time.Duration(oldest) * day).Unix()
	if _, err := ts.db.Exec(
		`INSERT INTO sku (source, id, name, first_seen, last_seen) VALUES (?, ?, 'Producto', ?, ?)`,
		source, id, first, ts.clock.Unix(),
	); err != nil {
		t.Fatal(err)
	}
	for daysAgo, price := range points {
		at := ts.clock.Add(-time.Duration(daysAgo) * day).Unix()
		if _, err := ts.db.Exec(
			`INSERT INTO observation (source, id, ts, price) VALUES (?, ?, ?, ?)`, source, id, at, price,
		); err != nil {
			t.Fatal(err)
		}
	}
}

func daily(from, to int, price float64) map[int]float64 {
	m := map[int]float64{}
	for d := from; d >= to; d-- {
		m[d] = price
	}
	return m
}

func merge(maps ...map[int]float64) map[int]float64 {
	out := map[int]float64{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func summaryFor(t *testing.T, ts *testStore, price float64) (products.PriceHistory, bool) {
	t.Helper()
	key := Key{Source: "disco", ID: "1", Price: price}
	result, err := ts.Summaries([]Key{key})
	if err != nil {
		t.Fatal(err)
	}
	summary, ok := result[key]
	return summary, ok
}

func TestSummaryNeedsEnoughPointsAndTime(t *testing.T) {
	useDefaults(t)

	few := newTestStore(t)
	seed(t, few, "disco", "1", map[int]float64{30: 1000, 20: 1000, 10: 1000, 0: 1000})
	if _, ok := summaryFor(t, few, 1000); ok {
		t.Fatal("4 observations should not be enough")
	}

	young := newTestStore(t)
	seed(t, young, "disco", "1", daily(10, 0, 1000))
	if _, ok := summaryFor(t, young, 1000); ok {
		t.Fatal("10 days of data should not be enough")
	}
}

func TestSteadyPriceHasNoChange(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	seed(t, ts, "disco", "1", daily(40, 0, 1000))

	summary, ok := summaryFor(t, ts, 1000)
	if !ok {
		t.Fatal("expected a summary")
	}
	if summary.ChangePct != 0 || summary.Typical != 1000 || summary.Low != 1000 || summary.High != 1000 {
		t.Fatalf("unexpected summary %+v", summary)
	}
	if summary.ChangeDays != 30 || summary.Points != 41 {
		t.Fatalf("expected a 30 day comparison over 41 points, got %+v", summary)
	}
}

func TestRecentDropIsReportedAgainstAMonthAgo(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	// 1000 until 6 days ago, 800 since then.
	seed(t, ts, "disco", "1", merge(daily(40, 6, 1000), daily(5, 0, 800)))

	summary, ok := summaryFor(t, ts, 800)
	if !ok {
		t.Fatal("expected a summary")
	}
	if summary.ChangePct != -20 {
		t.Fatalf("expected -20%% versus 30 days ago, got %v", summary.ChangePct)
	}
	if summary.Low != 800 || summary.High != 1000 || summary.Typical != 1000 {
		t.Fatalf("unexpected summary %+v", summary)
	}
}

func TestDaysWithoutReadingsCarryThePriceForward(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	// Only 5 readings in 40 days: the price changes on day 20, the rest are heartbeats.
	seed(t, ts, "disco", "1", map[int]float64{40: 1000, 30: 1000, 21: 1000, 20: 1200, 0: 1200})

	summary, ok := summaryFor(t, ts, 1200)
	if !ok {
		t.Fatal("expected a summary")
	}
	// 30 days ago the carried price was 1000, so the change is +20%.
	if summary.ChangePct != 20 || summary.Points != 5 {
		t.Fatalf("unexpected summary %+v", summary)
	}
}

func TestLivePriceIsUsedForToday(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	seed(t, ts, "disco", "1", daily(40, 0, 1000))

	summary, ok := summaryFor(t, ts, 900)
	if !ok {
		t.Fatal("expected a summary")
	}
	if summary.ChangePct != -10 || summary.Low != 900 {
		t.Fatalf("expected the live price to count, got %+v", summary)
	}
}

func TestShortHistoryComparesOverWhatItHas(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	seed(t, ts, "disco", "1", merge(daily(20, 4, 1000), daily(3, 0, 1100)))

	summary, ok := summaryFor(t, ts, 1100)
	if !ok {
		t.Fatal("expected a summary")
	}
	if summary.ChangeDays != 20 || summary.ChangePct != 10 {
		t.Fatalf("expected +10%% over 20 days, got %+v", summary)
	}
}

func TestUnknownProductHasNoSummary(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)

	result, err := ts.Summaries([]Key{{Source: "disco", ID: "nope", Price: 10}})
	if err != nil || len(result) != 0 {
		t.Fatalf("expected an empty result, got %v (%v)", result, err)
	}
}

func TestSeriesReturnsOnePricePerDay(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	seed(t, ts, "disco", "1", map[int]float64{9: 1000, 5: 1100, 0: 1100})

	name, series, ok, err := ts.Series("disco", "1", 10)
	if err != nil || !ok || name != "Producto" {
		t.Fatalf("unexpected result: %q %v %v", name, ok, err)
	}
	if len(series) != 10 {
		t.Fatalf("expected 10 days, got %d", len(series))
	}
	if series[0].Price != 1000 || series[len(series)-1].Price != 1100 {
		t.Fatalf("unexpected series %+v", series)
	}
	if series[len(series)-1].Day != "2026-10-01" {
		t.Fatalf("expected the last day to be the clock's day, got %s", series[len(series)-1].Day)
	}

	if _, _, ok, _ := ts.Series("disco", "missing", 10); ok {
		t.Fatal("expected an unknown product to be reported as such")
	}
}

func TestAnnotateAddsHistoryOnlyWhereThereIsEnough(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	seed(t, ts, "disco", "1", daily(40, 0, 1000))
	seed(t, ts, "disco", "2", daily(3, 0, 500))
	UseStore(ts.Store)
	t.Cleanup(func() { recorder = nil })

	list := []products.Schema{
		{Source: "disco", ID: "1", Price: 1000},
		{Source: "disco", ID: "2", Price: 500},
		{Source: "vea", ID: "3", Price: 10},
	}
	Annotate(list)

	if list[0].History == nil || list[1].History != nil || list[2].History != nil {
		t.Fatalf("unexpected annotations: %+v %+v %+v", list[0].History, list[1].History, list[2].History)
	}
}

func TestChangeIsMeasuredAgainstThePriceExactlyThirtyDaysAgo(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	// The price was 900 until 36 days ago, then 1000. A comparison against the start of the
	// data would say +11%, but 30 days ago it was already 1000.
	seed(t, ts, "disco", "1", merge(daily(40, 36, 900), daily(35, 0, 1000)))

	summary, ok := summaryFor(t, ts, 1000)
	if !ok {
		t.Fatal("expected a summary")
	}
	if summary.ChangePct != 0 || summary.ChangeDays != 30 {
		t.Fatalf("expected no change over 30 days, got %+v", summary)
	}
}

func TestPriceFromBeforeTheWindowIsCarriedIn(t *testing.T) {
	useDefaults(t)
	ts := newTestStore(t)
	// The only readings are 40 days ago and today. A 10 day window has none of its own at the
	// start, so the first day must carry the older price.
	seed(t, ts, "disco", "1", map[int]float64{40: 1000, 0: 1200})

	_, series, ok, err := ts.Series("disco", "1", 10)
	if err != nil || !ok {
		t.Fatalf("unexpected result: %v %v", ok, err)
	}
	if len(series) != 10 || series[0].Price != 1000 || series[8].Price != 1000 || series[9].Price != 1200 {
		t.Fatalf("expected the old price carried until today's change, got %+v", series)
	}
}
