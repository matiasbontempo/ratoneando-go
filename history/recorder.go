package history

import (
	"fmt"
	"strings"
	"time"

	"ratoneando/config"
	"ratoneando/products"
	"ratoneando/utils/logger"
)

const (
	backupInterval     = 24 * time.Hour
	startupBackupDelay = time.Minute
)

var recorder *Recorder

// Recorder takes readings off the request path and writes them from a single goroutine.
type Recorder struct {
	store *Store
	queue chan []Item
}

// Init starts recording if HISTORY_ENABLED is set. If anything fails, the API keeps working
// without history.
func Init() {
	if !config.HISTORY_ENABLED {
		logger.Log("Price history is disabled")
		return
	}

	store, err := Open(config.HISTORY_DB_PATH)
	if err != nil {
		logger.LogError("Price history disabled, could not open the database: " + err.Error())
		return
	}

	purgeConfiguredSources(store)

	start(store)
	go recorder.runBackups()

	logger.Log("Price history enabled: " + config.HISTORY_DB_PATH)
}

// purgeConfiguredSources runs the one-off cleanup asked for with HISTORY_PURGE_SOURCES.
func purgeConfiguredSources(store *Store) {
	var sources []string
	for _, source := range strings.Split(config.HISTORY_PURGE_SOURCES, ",") {
		if source = strings.TrimSpace(source); source != "" {
			sources = append(sources, source)
		}
	}
	if len(sources) == 0 {
		return
	}

	observations, skus, err := store.PurgeSources(sources)
	if err != nil {
		logger.LogError("Price history purge failed: " + err.Error())
		return
	}
	logger.Log(fmt.Sprintf(
		"Price history purged for %s: %d products, %d readings. Unset HISTORY_PURGE_SOURCES.",
		strings.Join(sources, ", "), skus, observations,
	))
}

// start installs the store and starts the writer goroutine. Tests use it with a temporary store.
func start(store *Store) {
	recorder = &Recorder{store: store, queue: make(chan []Item, 64)}
	go recorder.run()
}

// UseStore enables history on an already opened store, without a database path or backups.
// It exists for tests of the packages that read history.
func UseStore(store *Store) {
	start(store)
}

// Active returns the store when history is enabled, or nil.
func Active() *Store {
	if recorder == nil {
		return nil
	}
	return recorder.store
}

// Annotate adds a history summary to the products that have enough data. It never fails a
// search: on any error the products are left as they are.
func Annotate(shown []products.Schema) {
	store := Active()
	if store == nil || len(shown) == 0 {
		return
	}

	keys := make([]Key, len(shown))
	for i, product := range shown {
		keys[i] = Key{Source: product.Source, ID: product.ID, Price: product.Price}
	}

	summaries, err := store.Summaries(keys)
	if err != nil {
		logger.LogError("Price history read failed: " + err.Error())
		return
	}
	for i, key := range keys {
		if summary, ok := summaries[key]; ok {
			s := summary
			shown[i].History = &s
		}
	}
}

func (r *Recorder) run() {
	for items := range r.queue {
		result, err := r.store.Record(items)
		if err != nil {
			logger.LogError("Price history write failed: " + err.Error())
			continue
		}
		logger.LogDebug(fmt.Sprintf(
			"Price history: %d new, %d unchanged, %d skipped, %d suspicious",
			result.Inserted, result.Unchanged, result.Skipped, result.Suspicious,
		))
	}
}

func (r *Recorder) runBackups() {
	backup := func() {
		if err := r.store.Backup(); err != nil {
			logger.LogError("Price history backup failed: " + err.Error())
		}
	}

	// A first copy shortly after start, so there is a consistent file to download even if the
	// service is redeployed more often than once a day.
	time.Sleep(startupBackupDelay)
	backup()

	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()
	for range ticker.C {
		backup()
	}
}

// Record queues the products a user was shown. It never blocks: if the queue is full the
// readings are dropped, because history must not slow a search down.
func Record(shown []products.Schema) {
	if recorder == nil || len(shown) == 0 {
		return
	}

	items := make([]Item, 0, len(shown))
	for _, product := range shown {
		items = append(items, Item{
			Source:     product.Source,
			ID:         product.ID,
			Name:       product.Name,
			Link:       product.Link,
			Image:      product.Image,
			Unit:       product.Unit,
			UnitFactor: product.UnitFactor,
			Price:      product.Price,
			ListPrice:  product.ListPrice,
		})
	}

	select {
	case recorder.queue <- items:
	default:
		logger.LogWarn("Price history queue is full, dropping a batch")
	}
}
