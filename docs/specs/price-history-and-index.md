# Spec: Price History and the Ratoneando Inflation Index

Status (2026-10-07):
- **Recording is live in production** (PR #9, #10): SQLite on a Railway volume at `/data`,
  `HISTORY_ENABLED=true`.
- **Serving history is built but not merged**: search summaries and `GET /history` (Go branch
  `feat/history-api`), and the UI wired to it behind `VITE_PRICE_HISTORY` (web branch
  `feat/price-history-api`).
- The index (Part B) is not built. The basket needs the unit-price audit first.

Items marked **[Decision]** need an answer from the owner. Items marked **[Estimate]** are rough
guesses, not measurements.

---

## 1. Summary

Today Ratoneando answers one question: "what does this cost right now, across stores?". Every
search scrapes the stores live and the results are thrown away after a cache expires.

This spec adds memory:

1. **Price history.** Record the prices the app already sees, per store product (SKU), and show
   users whether a price is unusually low or high compared with recent weeks.
2. **Ratoneando Inflation Index.** A monthly index built on a fixed basket of staple products, with
   two series: the average shopper and the "ratón" (cheapest store per item).

Price history is the foundation. The index reuses its storage and its collection job.

### Goals
- Start accumulating data as early as possible. History cannot be backfilled.
- Keep infrastructure cost close to zero: it must run inside the current $5 Railway plan.
- Never slow down or break search. History is an enhancement that fails open.
- Be honest in the UI about what the numbers do and do not include.

### Non-goals (for now)
- Price alerts and push notifications.
- Matching the same product across stores (EAN), or "concept" level products.
- Card, wallet and in-store promotions.
- Other countries.
- Anything user-specific (accounts, saved lists).

---

## 2. Current system (relevant facts)

- **Realtime lookup.** `GET /?q=` runs 8 scrapers in parallel (`controllers/normalized_scrapper.go`),
  filters with a fuzzy match, sorts by unit price, and returns `products`, `failedScrapers` and
  `timestamp`.
- **Cache.** The whole response is cached in Redis by query string (`REDIS_CACHE_EXPIRATION`,
  default 28800 s, 8 hours). Nothing else is persisted. Redis has a 5 GB volume.
- **Product shape.** `products.Schema` has `id`, `source`, `name`, `link`, `image`, `price`, `unit`,
  `unitPrice`. The richer `ExtendedSchema` also has `listPrice`, `unavailable`, `unitFactor`, but
  `unit.CalculateUnitInfo` drops `listPrice` and `unitFactor` when it converts to `Schema`.
- **Unit price.** If no unit can be extracted, `unitPrice` falls back to the pack `price`
  (`unit/calculator.go`). So "unit price" is sometimes really pack price. The index has to handle that.
- **Unavailable products** are skipped in `cores/api/main.go` and leave zero-valued entries in the
  slice, which the fuzzy filter then drops.
- **Hosting.** One Railway service for the API, one for Redis, $5 plan. The VTEX query hash can
  go stale (see `docs/runbooks/rotate-vtex-hash.md`).

---

## 3. Principles

1. **Track SKUs, not concepts.** A SKU is `(source, id)`, the store's own identifier. It is exact
   and stable. A "concept" such as "leche entera 1 L" is a fuzzy grouping and stays out of the
   storage layer. Concepts become a view on top of SKUs later.
2. **Store raw inputs, derive the rest.** Unit extraction changes over time. Store the product name
   and the price as the store reported them, plus the extractor's output, so numbers can be
   recomputed.
3. **History never blocks a search.** Recording is asynchronous. If the database is down, the search
   response simply omits history fields.
4. **Say less when unsure.** Show nothing when there are too few observations. Do not use words like
   "historic" without months of data behind them.
5. **Flag everything.** Backend recording and the UI each ship behind a flag, off by default.

---

## 4. Part A: Price history

### 4.1 Data model

SQLite, two tables.

```sql
CREATE TABLE sku (
  source      TEXT NOT NULL,
  id          TEXT NOT NULL,
  name        TEXT NOT NULL,        -- latest raw name, kept to re-run unit extraction
  link        TEXT,
  image       TEXT,
  unit        TEXT,                 -- extractor output, latest
  unit_factor REAL,                 -- extractor output, latest (needs plumbing, see 4.2)
  first_seen  INTEGER NOT NULL,     -- unix seconds
  last_seen   INTEGER NOT NULL,
  PRIMARY KEY (source, id)
);

CREATE TABLE observation (
  source     TEXT NOT NULL,
  id         TEXT NOT NULL,
  ts         INTEGER NOT NULL,      -- unix seconds
  price      REAL NOT NULL,
  list_price REAL,                  -- nullable, needs plumbing
  PRIMARY KEY (source, id, ts)
) WITHOUT ROWID;
```

No user identifiers, IP addresses or request metadata are stored.

### 4.2 Ingestion

**Where.** In `NormalizedScraper`, after the response is built on a cache miss, hand the filtered
products to a goroutine. Do not wait for it.

**What gets recorded.** Only products that passed the fuzzy filter, so unrelated results never enter
the database.

**Rules per product**
- Skip empty `id` (the zero-valued entries left by unavailable products), `price <= 0`.
- Hold back a reading that is more than 3x above or below the last stored price (a "sanity bound",
  to keep a glitchy `$1` or `$900,000` out of the history). If the same odd price (within 2%) is
  seen 3 times, it is accepted as real; a normal reading in between resets the count. Implemented
  in `history/store.go`.
- Insert a row only if `price` or `list_price` changed since the last stored row, or if the last
  row is older than 24 hours (a heartbeat, so "unchanged" can be told apart from "not seen").
- Upsert `sku` with the latest name, link, image, unit, `last_seen`.
- The "last stored price" comes from one indexed read per product in SQLite (the primary key is
  `(source, id, ts)`). A Redis cache for it was dropped as unnecessary, and it keeps history
  independent of the Redis flushes in the cache workflow.

**Plumbing required.** `products.Schema` needs `ListPrice` and the unit factor carried through
`CalculateUnitInfo`. Both are internal; the public response only gains the history fields in 4.4.

**Search-driven only.** There is no scheduled refresh of popular queries. Popular queries already
refresh themselves: the response cache lasts 8 hours, so a busy query is re-scraped (and recorded)
up to three times a day by real traffic. A scheduled job would only add products that are searched
less than once a day, which are the least important ones. The consequence is that a product's
series is dense where users look and sparse elsewhere, which is the right bias for badges.

The one exception is the index: a fixed basket cannot depend on user traffic. A small scheduled job
for the basket queries is part of Part B (about 40 items across the supermarket chains, a few hundred
requests a night **[Estimate]**).

**Catalog by-product.** Because every recorded SKU keeps its name, link and image, the `sku` table
becomes a growing catalog of the products users actually look for. That is the seed for a future
in-house search or autocomplete, and for showing the last known price when a store's scraper is down.
It is a skewed catalog (what was searched), not a full crawl of the stores.

**Search terms** are not stored by this feature. PostHog already records every search
(`search_product`), and the owner can supply the top terms for choosing the basket.

**Failure behavior.** If recording fails, log and move on. If the database is unavailable at
startup, the API starts with history disabled.

### 4.3 Storage and operations

- **Engine:** SQLite via a pure-Go driver (`modernc.org/sqlite`), so the build stays CGO-free.
  WAL mode, single writer goroutine fed by a channel.
- **Where:** a Railway volume attached to the API service (mounted outside `/app`, e.g. `/data`).
  Volumes are mounted when the container starts, not at build time.
- **Backup (decided):** Railway's scheduled volume backups are **Pro-plan only**, so they are not
  available on the current Hobby plan. Instead the app writes a consistent copy with `VACUUM INTO`
  to `<db>.backup` one minute after every start and then every 24 hours, and the owner downloads it
  by hand: `railway volume files download /history.db.backup ./history-backup.db`. That copy sits
  on the same volume, so it protects against corruption, not against losing the volume, which is
  why the manual download matters. **Wiping a volume deletes it and any platform backups**, so
  never wipe it. A daily upload to Cloudflare R2, or Litestream, remains an option if manual
  downloads become a burden.
- **Size [Estimate]:** an observation is tens of bytes. 10,000 tracked SKUs changing roughly weekly
  is about 500,000 rows per year, around 30 MB. Railway bills volumes by the storage actually
  used (documented at $0.15 per GB per month), so this is pennies.
- **Trade-off:** one writer only, so no horizontal scaling of the API. I also believe a service
  with a volume gets a short downtime window on each deploy, but the docs pages I read do not
  confirm this, so test it.
- **Railway's database options.** The Railway template list offers PostgreSQL, Redis, MongoDB and
  MySQL. SQLite is not one of them: it is just a file on a volume.
  - **PostgreSQL** is the real alternative. No deploy-downtime question and richer queries for the
    index, but it is an extra service whose CPU and memory count against the plan's included usage.
    The Hobby plan includes $5 of usage and bills only the excess, so the question is how much
    headroom the account has today. That has not been measured.
  - **Redis** is ruled out for history. It is already used as a cache that gets flushed
    (`FLUSHDB` is part of the hash-rotation runbook), which would wipe the data.
  - **MongoDB and MySQL** offer no advantage over the two options above.
- **Retention:** keep everything for now. At the estimated size it is not a problem.

### 4.4 API

**Search response (implemented).** Each product may gain an optional `history` summary, computed
when the response is built (so it is cached with the response, up to 8 hours stale):

```json
{
  "id": "676752", "source": "carrefour", "price": 1939, "...": "...",
  "history": {
    "points": 41,
    "windowDays": 46,
    "typical": 2272,
    "low": 1799,
    "high": 2392,
    "changePct": -13.2,
    "changeDays": 30
  }
}
```

Definitions:
- The series is one price per UTC day over the last `HISTORY_WINDOW_DAYS` (default 60), carrying
  the last known price forward across days with no new reading. Days before the first known
  price are left out, so `windowDays` is the number of days actually covered.
- Today's point uses the price being shown right now, not the last stored one.
- `typical`: median of the daily series. `low` / `high`: min and max.
- `changePct`: change in percent between today and `changeDays` days ago, where `changeDays` is
  `min(30, days since the product was first seen)`. A young history therefore compares over a
  shorter period and says so.
- `points`: stored observations inside the window.
- The object is **omitted** unless there are at least `HISTORY_MIN_POINTS` (default 5)
  observations and the product was first seen at least `HISTORY_MIN_SPAN_DAYS` (default 14) days ago.
- The summary is read through a separate query-only connection pool; any error leaves the product
  without a summary and never fails the search.

**History endpoint (implemented).** `GET /history?source=&id=[&days=]` returns
`{ "source", "id", "name", "series": [{ "d": "2026-09-01", "p": 1850 }, ...] }`, `days` from 2 to
180 (default the window). It sends `Cache-Control: public, max-age=3600`, applies the same referer
rule as search in release mode, and answers 400 for bad parameters and 404 for unknown products or
when history is disabled. The web app fetches it only when the detail sheet opens.

### 4.5 User experience

- Search itself is unchanged: same input, same spinner, same order (unit price by default).
- **Badges.** A product shows a small badge only when it is worth mentioning:
  - at or near its lowest price in the window and clearly below typical: "Precio más bajo en N semanas";
  - price change of at least 8%: "▼ X% vs. el último mes" / "▲ X% este mes" for a full month of
    comparison, or "▼ X% en N días" / "▲ X% en N días" while the history is younger.
- **At most 3 badges per search**, ranked by size of the move; "lowest in weeks" outranks a plain drop.
- **Detail sheet** on tapping a badge: chart (about 60 days), low / typical / high, a note that
  prices exclude card, wallet and in-store promotions, and a link to the store.
- **Opt-in sort "Mejores ofertas"** may be added later. The default order never changes.
- **Inflation caveat.** In a high-inflation market a long-window "typical price" makes almost
  everything look expensive. The badge therefore uses a 30-day comparison, and wording avoids
  "histórico". A later improvement is to subtract the market-wide trend (see Part B).

### 4.6 Flags and configuration

- Backend: `HISTORY_ENABLED` (record and serve), `HISTORY_DB_PATH`, `HISTORY_WINDOW_DAYS` (60),
  `HISTORY_MIN_POINTS` (5), `HISTORY_MIN_SPAN_DAYS` (14), `PARTIAL_CACHE_EXPIRATION` (60 s, how long
  a response with failed stores stays cached).
- Frontend: `VITE_PRICE_HISTORY` shows badges and the sheet from API data; with it off the UI
  ignores the `history` objects the API sends. `VITE_PRICE_HISTORY_MOCK` replaces them with fake
  data for design work.

### 4.7 Metrics (via the existing PostHog setup)

- Badge impressions per search; taps on a badge (`results_open_price_history`, already added in the mock).
- Add-to-cart rate for badged vs. non-badged products.
- Return visits for users who opened a sheet.
- If sheet opens are close to zero after a few weeks, treat the feature as decoration.

---

## 5. Part B: Ratoneando Inflation Index

### 5.1 Definition

A monthly chain-linked index with base 100, built from the **same SKUs** compared month to month.
Two series are published:

- **Average basket:** how prices moved across all stores.
- **Ratón basket:** the cost of buying each item at the cheapest store each month.
- Derived stat: the gap between the two ("comparing saves X%").

It is named "Índice Ratoneando" and never presented as official.

### 5.2 Basket

- About **40 staple items** **[Decision: the owner picks, or the assistant proposes from the owner's list of top search terms]**,
  e.g. milk, oil, rice, pasta, sugar, yerba, flour, eggs.
- Each item is a fixed specification: a product type and a pack size ("leche entera larga vida 1 L").
- For each store, each item is **frozen to one SKU** chosen at the base date. **[Decision: how to
  choose. Proposed: the first result for a fixed query in that store, reviewed by hand once.]**
  Freezing prevents the sample from drifting with popularity.
- A store may lack an item. An item counts only if at least **3 of the 6 supermarket chains** carry it. The index uses only the
  chains the app currently shows to users (Carrefour, Disco, Jumbo, Vea, Día, MasOnline). Farmacity
  and MercadoLibre are excluded: a pharmacy and a marketplace are not comparable shelves.
- If a frozen SKU disappears for more than 14 days, replace it with the closest equivalent in that
  store and record the replacement date. The replacement is linked into the series (its first
  price is compared with the old SKU's last price only as a note, not as a price change).

### 5.3 Measuring

- **Price used:** unit price, so shrinkflation shows up as inflation. Items without a recognizable
  unit fall back to pack price, which is the existing behavior of `unitPrice`. The basket should
  include only items where the unit is reliable, and a validation step flags the rest.
- **Item-store monthly price** `P(i,s,m)`: median of the daily unit prices of that SKU in month `m`.
  Medians make the monthly value robust to a short promotion.
- **Item price (average basket)** `P(i,m)`: geometric mean across the stores that carry item `i` in
  both `m` and `m-1`.
- **Item price (ratón basket)** `P*(i,m)`: minimum across stores of `P(i,s,m)`.
- **Price relative** `r(i,m) = P(i,m) / P(i,m-1)`, over items present in both months.
- **Index:** `I(m) = I(m-1) × geomean_i( r(i,m) )`, with `I(base) = 100`. Same formula with
  `P*` for the ratón series. Equal weights.
- **Outliers:** drop a relative outside **[Decision: 0.5 to 2.0]** from the month's calculation and
  list it in the audit output.

### 5.4 Publication rules

- Collect daily, compute and publish monthly.
- Publish a month only if at least 80% of basket items have valid prices.
- Do not publish before roughly 8 to 12 weeks of daily history exist.
- Publish alongside: methodology page, the basket list, number of items used, and the INDEC
  monthly food and beverages figure for comparison. A persistent large gap needs an explanation
  before publishing. A gap by itself is not proof of an error.
- The first public release should be reviewed by the owner against raw data.

### 5.5 Known limitations (stated in the methodology)

- Online shelf prices only: no card, wallet or in-store discounts.
- Online prices may differ from branch prices, and by region. Region handling is not verified.
- Equal weights are not a household spending pattern. Weighting by INDEC expenditure shares is a
  possible later improvement.
- Six chains are not the whole market.

### 5.6 Implementation notes

- A `basket` table (`item`, `source`, `sku_id`, `valid_from`, `valid_to`) records which SKU stands
  for which item over time. The index job reads `observation` and `basket`; nothing else.
- A command, `cmd/compute_index`, outputs the series and an audit report (which SKUs, which
  months, which outliers). It runs by hand at first.

---

## 6. Phases

| Phase | Scope | Exit criteria |
|---|---|---|
| 1. Collection | Schema, recording on cache miss, change-only logic, flag, volume + scheduled backup | Observations accumulating for a week with no effect on latency or errors |
| 2. Serve history | `history` fields in search, `/history` endpoint, real UI flag | Badges appear with real data; sheet chart loads lazily |
| 3. Basket | Unit-price audit, choose items and SKUs, `basket` table, scheduled basket job, `cmd/compute_index`, audit report | First month of internal numbers reviewed by hand |
| 4. Publish | Methodology page, public index and ratón series, INDEC comparison | Owner sign-off |

**[Estimate]** Phase 1 a few days, phase 2 several days (including the
move from mock to API), phases 3 and 4 depend mostly on decisions and on waiting for data.
Calendar time is dominated by data accumulation, not coding.

---

## 7. Risks

| Risk | Mitigation |
|---|---|
| A scraper returns wrong prices and they get stored | Sanity bounds, change audit log, pause the basket job when most stores fail, manual review of outliers |
| A store changes its product IDs | Monitor sudden mass appearance and disappearance of SKUs per store |
| Stale VTEX hash silently stops collection | Existing runbook, plus a canary and alert (not yet scheduled) |
| Deploy downtime from the volume (unconfirmed) | Test it, accept it, or move to Postgres |
| Volume wiped by mistake (this deletes its backups too) | Treat the volume as untouchable; keep the nightly `VACUUM INTO` copy and consider an off-platform copy |
| Redis and SQLite disagree on "last price" | Redis is only a cache; SQLite is the source of truth, rebuilt on miss |
| Coverage is biased toward what users search | The index uses a fixed basket with its own scheduled job, not search traffic; history is only shown where there is enough data |
| Credibility of the index | Public methodology, audit trail, comparison with INDEC, internal-only until reviewed |
| Legal and terms-of-service exposure from collecting and republishing store prices | Not assessed in this spec. Review before the public launch |
| Region and branch effects on prices | Not verified. Check whether prices vary by postal code before publishing the index |

---

## 8. Open decisions

1. Basket: the owner will provide the most common search terms; the assistant proposes items from that list after the unit-price audit.
2. How to choose the frozen SKU per store for each item.
3. Sanity bounds for the index outliers (0.5 to 2.0?). The ingestion bound is decided and implemented (3x, confirmed after 3 sightings).
4. ~~SQLite vs. PostgreSQL~~ **Decided: SQLite on a Railway volume with scheduled volume backups.**
   PostgreSQL stays the fallback if deploy downtime or backups turn out to be a problem.
5. Does the "ratón" series lead the public messaging?
6. Does every product get a tappable "ver historial", or only the badged ones?

---

## 9. Later

Alerts ("avisame si baja de $X"), EAN-based cross-store matching and canonical products, fake-discount
detection using `list_price`, promotions and card discounts, per-concept price charts for a query,
shareable lists through the URL shortener, and other countries.
