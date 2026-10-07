package history

import (
	"fmt"
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

	recorder = &Recorder{store: store, queue: make(chan []Item, 64)}
	go recorder.run()
	go recorder.runBackups()

	logger.Log("Price history enabled: " + config.HISTORY_DB_PATH)
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
