package history

import (
	"database/sql"
	"math"
	"sort"
	"time"

	"ratoneando/config"
	"ratoneando/products"
)

const day = 24 * time.Hour

// Key identifies a product in a store, together with the price the user is being shown today.
type Key struct {
	Source string
	ID     string
	Price  float64
}

// DayPrice is the price of a product on one day (UTC).
type DayPrice struct {
	Day   string  `json:"d"`
	Price float64 `json:"p"`
}

type observed struct {
	ts    int64
	price float64
}

func startOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// dailySeries turns observations into one price per day, carrying the last known price forward
// across days with no new reading. Days before the first known price are left out.
// baseline is the last price before the window (0 if there was none). If livePrice is positive
// it replaces the price of the last day.
func dailySeries(window []observed, baseline float64, from time.Time, days int, livePrice float64) []DayPrice {
	series := make([]DayPrice, 0, days)
	price := baseline
	next := 0

	for i := 0; i < days; i++ {
		date := from.Add(time.Duration(i) * day)
		end := date.Add(day).Unix() - 1
		for next < len(window) && window[next].ts <= end {
			price = window[next].price
			next++
		}
		if i == days-1 && livePrice > 0 {
			price = livePrice
		}
		if price > 0 {
			series = append(series, DayPrice{Day: date.Format("2006-01-02"), Price: price})
		}
	}
	return series
}

func median(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

func round(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

type queryer interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

func loadSeries(q queryer, source, id string, from time.Time, days int, livePrice float64) ([]DayPrice, []observed, int64, error) {
	var firstSeen int64
	if err := q.QueryRow(`SELECT first_seen FROM sku WHERE source = ? AND id = ?`, source, id).Scan(&firstSeen); err != nil {
		return nil, nil, 0, err
	}

	var baseline float64
	err := q.QueryRow(
		`SELECT price FROM observation WHERE source = ? AND id = ? AND ts < ? ORDER BY ts DESC LIMIT 1`,
		source, id, from.Unix(),
	).Scan(&baseline)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, 0, err
	}

	rows, err := q.Query(
		`SELECT ts, price FROM observation WHERE source = ? AND id = ? AND ts >= ? ORDER BY ts`,
		source, id, from.Unix(),
	)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var window []observed
	for rows.Next() {
		var o observed
		if err := rows.Scan(&o.ts, &o.price); err != nil {
			return nil, nil, 0, err
		}
		window = append(window, o)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, 0, err
	}

	return dailySeries(window, baseline, from, days, livePrice), window, firstSeen, nil
}

// summarize decides whether there is enough data and, if so, builds the summary.
func summarize(series []DayPrice, points int, firstSeen int64, now time.Time) (products.PriceHistory, bool) {
	spanDays := int(now.Sub(time.Unix(firstSeen, 0)) / day)
	if points < config.HISTORY_MIN_POINTS || spanDays < config.HISTORY_MIN_SPAN_DAYS || len(series) < 2 {
		return products.PriceHistory{}, false
	}

	prices := make([]float64, len(series))
	low, high := math.Inf(1), math.Inf(-1)
	for i, p := range series {
		prices[i] = p.Price
		low = math.Min(low, p.Price)
		high = math.Max(high, p.Price)
	}

	changeDays := spanDays
	if changeDays > 30 {
		changeDays = 30
	}
	if changeDays > len(series)-1 {
		changeDays = len(series) - 1
	}
	reference := series[len(series)-1-changeDays].Price
	current := series[len(series)-1].Price

	return products.PriceHistory{
		Points:     points,
		WindowDays: len(series),
		Typical:    round(median(prices), 0),
		Low:        low,
		High:       high,
		ChangePct:  round((current-reference)/reference*100, 1),
		ChangeDays: changeDays,
	}, true
}

// Summaries returns a summary for each product that has enough history. Products without
// enough data are simply absent from the result.
func (s *Store) Summaries(keys []Key) (map[Key]products.PriceHistory, error) {
	now := s.now()
	days := config.HISTORY_WINDOW_DAYS
	from := startOfDay(now).Add(-time.Duration(days-1) * day)

	tx, err := s.readDB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result := map[Key]products.PriceHistory{}
	for _, key := range keys {
		series, window, firstSeen, err := loadSeries(tx, key.Source, key.ID, from, days, key.Price)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		if summary, ok := summarize(series, len(window), firstSeen, now); ok {
			result[key] = summary
		}
	}
	return result, nil
}

// Series returns the daily prices of one product, or ok=false if the product is unknown.
func (s *Store) Series(source, id string, days int) (name string, series []DayPrice, ok bool, err error) {
	now := s.now()
	from := startOfDay(now).Add(-time.Duration(days-1) * day)

	err = s.readDB.QueryRow(`SELECT name FROM sku WHERE source = ? AND id = ?`, source, id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, err
	}

	series, _, _, err = loadSeries(s.readDB, source, id, from, days, 0)
	if err != nil {
		return "", nil, false, err
	}
	return name, series, true, nil
}
