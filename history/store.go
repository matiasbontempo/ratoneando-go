package history

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	// A product that has not changed price is still recorded once a day, so a gap in the series
	// means "not seen" and not "unchanged".
	heartbeat = 24 * time.Hour

	// A reading that is more than this factor above or below the last stored price is held back
	// as suspicious, unless it keeps repeating (see suspectConfirmations).
	maxPriceJump = 3.0

	// How many times the same odd price has to be seen before it is accepted as real.
	suspectConfirmations = 3

	// Two prices closer than this (relative) are considered the same suspicious price.
	suspectTolerance = 0.02
)

const schema = `
CREATE TABLE IF NOT EXISTS sku (
  source      TEXT NOT NULL,
  id          TEXT NOT NULL,
  name        TEXT NOT NULL,
  link        TEXT,
  image       TEXT,
  unit        TEXT,
  unit_factor REAL,
  first_seen  INTEGER NOT NULL,
  last_seen   INTEGER NOT NULL,
  PRIMARY KEY (source, id)
);

CREATE TABLE IF NOT EXISTS observation (
  source     TEXT NOT NULL,
  id         TEXT NOT NULL,
  ts         INTEGER NOT NULL,
  price      REAL NOT NULL,
  list_price REAL,
  PRIMARY KEY (source, id, ts)
) WITHOUT ROWID;
`

// Item is one product reading, as seen in a store at one moment.
type Item struct {
	Source     string
	ID         string
	Name       string
	Link       string
	Image      string
	Unit       string
	UnitFactor float64
	Price      float64
	ListPrice  float64
}

// Result tells what Record did, mostly for logs and tests.
type Result struct {
	Inserted   int
	Unchanged  int
	Skipped    int
	Suspicious int
}

type suspect struct {
	price float64
	count int
}

type Store struct {
	db       *sql.DB
	path     string
	now      func() time.Time
	mu       sync.Mutex
	suspects map[string]suspect
}

// Open opens (and creates if needed) the SQLite database at path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating history directory: %w", err)
	}

	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening history database: %w", err)
	}
	// SQLite allows a single writer. One connection avoids "database is locked" surprises.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating history schema: %w", err)
	}

	return &Store{
		db:       db,
		path:     path,
		now:      time.Now,
		suspects: map[string]suspect{},
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func valid(item Item) bool {
	return item.Source != "" &&
		item.ID != "" &&
		item.Price > 0 &&
		!math.IsNaN(item.Price) &&
		!math.IsInf(item.Price, 0)
}

func isJump(previous, current float64) bool {
	ratio := current / previous
	return ratio > maxPriceJump || ratio < 1/maxPriceJump
}

func closeTo(a, b float64) bool {
	return math.Abs(a-b) <= suspectTolerance*math.Max(a, b)
}

// Record stores the readings that carry new information. It is safe to call from several
// goroutines, though the Recorder normally serializes calls.
func (s *Store) Record(items []Item) (Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result Result
	now := s.now().Unix()

	tx, err := s.db.Begin()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	for _, item := range items {
		if !valid(item) {
			result.Skipped++
			continue
		}

		var lastPrice, lastList sql.NullFloat64
		var lastTs int64
		err := tx.QueryRow(
			`SELECT price, list_price, ts FROM observation
			 WHERE source = ? AND id = ? ORDER BY ts DESC LIMIT 1`,
			item.Source, item.ID,
		).Scan(&lastPrice, &lastList, &lastTs)
		hasLast := err == nil
		if err != nil && err != sql.ErrNoRows {
			return result, err
		}

		key := item.Source + "|" + item.ID
		if hasLast && isJump(lastPrice.Float64, item.Price) {
			seen := s.suspects[key]
			if closeTo(seen.price, item.Price) {
				seen.count++
			} else {
				seen = suspect{price: item.Price, count: 1}
			}
			s.suspects[key] = seen

			if seen.count < suspectConfirmations {
				result.Suspicious++
				continue
			}
		}
		delete(s.suspects, key)

		if _, err := tx.Exec(
			`INSERT INTO sku (source, id, name, link, image, unit, unit_factor, first_seen, last_seen)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (source, id) DO UPDATE SET
			   name = excluded.name, link = excluded.link, image = excluded.image,
			   unit = excluded.unit, unit_factor = excluded.unit_factor, last_seen = excluded.last_seen`,
			item.Source, item.ID, item.Name, item.Link, item.Image, item.Unit, item.UnitFactor, now, now,
		); err != nil {
			return result, err
		}

		changed := !hasLast ||
			lastPrice.Float64 != item.Price ||
			lastList.Float64 != item.ListPrice
		stale := hasLast && now-lastTs >= int64(heartbeat.Seconds())

		if !changed && !stale {
			result.Unchanged++
			continue
		}

		var list any
		if item.ListPrice > 0 {
			list = item.ListPrice
		}
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO observation (source, id, ts, price, list_price) VALUES (?, ?, ?, ?, ?)`,
			item.Source, item.ID, now, item.Price, list,
		); err != nil {
			return result, err
		}
		result.Inserted++
	}

	return result, tx.Commit()
}

// Backup writes a clean, consistent copy of the database next to it. Railway's volume backups
// then snapshot a file that is guaranteed to be a complete database.
func (s *Store) Backup() error {
	target := s.path + ".backup"
	tmp := target + ".tmp"

	if err := os.Remove(tmp); err != nil && !os.IsNotExist(err) {
		return err
	}
	if _, err := s.db.Exec("VACUUM INTO ?", tmp); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}
	return os.Rename(tmp, target)
}
