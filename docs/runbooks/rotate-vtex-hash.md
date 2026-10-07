# Runbook: rotate the VTEX persisted-query hash

VTEX stores (Carrefour, Disco, Vea, MasOnline, Farmacity, DiaOnline) are queried with a
persisted GraphQL query identified by a SHA256 hash. VTEX rotates it every few months. When it does,
every VTEX scraper stops returning results.

## 1. Recognize the problem

- Logs show `API returned error: PersistedQueryNotFound for <query>@<source>`.
- Carrefour, Disco, Vea, MasOnline, Farmacity and DiaOnline return empty or error results while
  non-VTEX scrapers (for example Jumbo and MercadoLibre) still work.
- `go run ./cmd/verify_vtex` prints `Hash validation FAILED`.

Quick check without the repo:

```bash
curl -s "https://www.carrefour.com.ar/_v/segment/graphql/v1/?workspace=master&maxAge=medium&appsEtag=remove&domain=store&locale=es-AR&operationName=productSuggestions&variables=%7B%7D&extensions=<URL-ENCODED-EXTENSIONS>" | head -c 300
```

`{"errors":[{"message":"PersistedQueryNotFound"...` means the hash is stale.

## 2. Get the new hash

1. Open https://www.carrefour.com.ar in Chrome, DevTools, Network tab.
2. Click the search box and type any product (for example `leche`). Clicking it first matters:
   programmatic focus does not trigger the suggestions request.
3. Find the request with `operationName=productSuggestions` (filter by `graphql`) and copy its full URL.
4. Extract the hash:

   ```bash
   go run ./cmd/decode_vtex     # paste the URL when prompted
   ```

   The output is a 64-character hex string.

Claude Code can do steps 1 to 3 through the Claude in Chrome extension.

## 3. Verify it locally

```bash
VTEX_SHA256_HASH=<new-hash> go run ./cmd/verify_vtex
```

Expect `Hash validation SUCCESSFUL!` and a product for the query. This only tests Carrefour. For a
fuller check, also run the Disco, Vea and other VTEX scrapers for a query like `leche`; each should
return products with no error.

The hash is read from the `VTEX_SHA256_HASH` env variable (`config/config.go`). The default is
`REPLACE_ME`, so locally set it in `.env` too.

## 4. Update production (Railway)

Project `ratoneando.ar`, environment `production`, service `ratoneando-go` (Redis is a separate
service in the same project). The Railway IDs are not stored in this repo. Find them with
`railway status` or in the dashboard URL, or ask Claude Code to list them through the Railway MCP.

Dashboard: set `VTEX_SHA256_HASH` on `ratoneando-go`.
CLI: `railway variables --set VTEX_SHA256_HASH=<new-hash> -s ratoneando-go`.
Claude Code with the Railway MCP: `set-variables` with only that one variable (no need to read the
other variables).

Changing a variable redeploys the service. The hash is read once at startup, so the redeploy is
required.

## 5. Flush the cache

Responses are cached in Redis (`RESPONSE_CACHE_EXPIRATION` 3600s, `REDIS_CACHE_EXPIRATION` 28800s by
default), so stale results can linger after the fix.

```bash
railway connect Redis        # opens redis-cli on the Railway Redis
> FLUSHDB
```

Notes:
- Redis has a persistent volume, so restarting it does not flush anything.
- The Railway MCP has no tool to run Redis commands. Reading the Redis URL through the MCP
  (`list-variables`) can be blocked by Claude Code's permission classifier because it exposes
  credentials. Run the flush yourself or allow that tool explicitly.
- If you skip the flush, entries expire on their own within 8 hours.

## 6. Confirm

1. Deployment status is `SUCCESS` (dashboard, or `list-deployments` with the MCP).
2. Hit the production API with a search, for example `leche`, and confirm results come from the
   VTEX stores.
3. Check the deploy logs for new `PersistedQueryNotFound` errors.

## History

- 2026-10-07: rotated to `1723eecb…bad` (stale `db333c9c…`). Hash set on Railway via MCP; cache flush pending.
