# Insights Astrology — Go SDK

Official Go client for the [Insights Astrology API](https://tuffys-ai-astrology.vercel.app) — a commercial astronomy API covering all 43 endpoints across Western, Vedic, Hellenistic, and electional traditions. Computed from first principles on a home-grown VSOP87D + ELP2000 engine.

**Zero runtime dependencies.** Only `net/http` and `encoding/json` from the standard library.

<p align="center">
  <a href="https://pkg.go.dev/github.com/omkarjaliparthi/tuffys-astrology-go/tuffys"><img src="https://img.shields.io/badge/pkg.go.dev-docs-007d9c?style=flat-square" /></a>
  <a href="https://tuffys-ai-astrology.vercel.app/docs/api"><img src="https://img.shields.io/badge/API_docs-Scalar-6E56CF?style=flat-square" /></a>
  <a href="https://tuffys-ai-astrology.vercel.app/pricing"><img src="https://img.shields.io/badge/Pricing-tiered-success?style=flat-square" /></a>
  <img src="https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square" />
  <img src="https://img.shields.io/badge/License-MIT-lightgrey?style=flat-square" />
  <img src="https://img.shields.io/badge/Dependencies-zero-brightgreen?style=flat-square" />
</p>

---

## Install

```bash
go get github.com/omkarjaliparthi/tuffys-astrology-go/tuffys@latest
```

Requires Go 1.21+. See [`SETUP.md`](./SETUP.md) for publishing internals.

---

## Quick start

```go
package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/omkarjaliparthi/tuffys-astrology-go/tuffys"
)

func main() {
    client := tuffys.New("https://tuffys-ai-astrology.vercel.app",
        tuffys.WithAPIKey("YOUR_JWT_HERE"),
    )

    chart, err := client.NatalChart(context.Background(), tuffys.Person{
        Datetime:  "1990-06-15T12:00:00Z", // strict ISO-8601, no naive locals
        Latitude:  51.5,
        Longitude: 0,
    })
    if err != nil {
        var apiErr *tuffys.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API %d %s: %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
            return
        }
        panic(err)
    }
    fmt.Printf("%+v\n", chart)
}
```

Errors surface as `*tuffys.APIError` with `Status`, `Code`, `Message`, `Details` — safe to unwrap with `errors.As`.

---

## Authentication

The API uses stateless HS256 JWTs. Each key encodes its own per-minute + per-day quotas — no round-trip to a key-store, horizontally scalable out of the box. Pass your key once at client construction:

```go
client := tuffys.New(baseURL, tuffys.WithAPIKey(os.Getenv("ASTROLOGY_API_KEY")))
```

See the [API docs](https://tuffys-ai-astrology.vercel.app/docs/api) for key minting + tier quotas.

---

## Endpoint coverage

All 43 v1 endpoints, grouped by domain:

| Domain | Endpoints |
|---|---|
| **Western** | `/chart/natal` · `/chart/extended` · `/positions` · `/houses` · `/aspects` · `/harmonic` · `/dignities` · `/asteroids` · `/fixed-stars` |
| **Vedic** | `/vedic/chart` · `/vedic/panchanga` · `/dashas` · `/chara-karakas` · `/yogas` · `/muhurta` · `/ashtakavarga` · `/shadbala` · `/kp-sublords` |
| **Hellenistic** | `/profections` · `/zodiacal-releasing` · `/planetary-nodes` · `/true-node` |
| **Relational** | `/transits` · `/synastry` · `/composite` · `/progressions` · `/returns/solar` · `/returns/lunar` |
| **Electional / events** | `/eclipses` · `/eclipses/prenatal` · `/planetary-hours` · `/voc-moon` · `/sun-rise-set` |
| **Aggregators** | `/daily` · `/compatibility` · `/events` |
| **Utility** | `/geocode` · `/openapi.json` |

Every method is typed end-to-end. No `interface{}`, no magic strings.

---

## Design principles

- **Zero runtime dependencies.** Embed this in anything — serverless, edge, CLI, bot — without dragging a graph.
- **Context-first.** Every method takes `context.Context`. Cancel anywhere in the tree.
- **Errors are data.** `*tuffys.APIError` carries the HTTP status, the machine-readable code, and a human message.
- **Idiomatic options.** Functional options (`WithAPIKey`, `WithHTTPClient`, `WithUserAgent`) — swap out the transport for tracing, retries, or testing.
- **Strict input parsing on the server.** ISO-8601 datetimes required — no ambiguous local times. Pair with Go's `time.Time.Format(time.RFC3339)` and you're set.

---

## Related

- **[Main repository](https://github.com/omkarjaliparthi/tuffys-ai-astrology)** — engine, API, frontend, case study
- **[TypeScript SDK](https://www.npmjs.com/package/tuffys-astrology)** — published on npm
- **[Python SDK](https://pypi.org/project/tuffys-astrology/)** — published on PyPI
- **[API docs](https://tuffys-ai-astrology.vercel.app/docs/api)** — interactive Scalar explorer
- **[Pricing](https://tuffys-ai-astrology.vercel.app/pricing)** — Developer (free) · Studio ($49/mo) · Scale (custom)

---

## License

MIT — see [`LICENSE`](./LICENSE). Free to embed in commercial applications.
