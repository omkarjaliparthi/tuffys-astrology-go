# Kriya — Go SDK

Official Go client for [**Kriya**](https://kriya.insightsbyomkar.com) — the Insights Astrology API by Insights by Omkar. A commercial astronomy API covering **109+ v1 endpoints** across Western, Vedic, Hellenistic, Jaimini, KP, and electional traditions. Computed from first principles on a home-grown VSOP87D + ELP2000 + DOPRI8 engine.

**Zero runtime dependencies.** Only `net/http` and `encoding/json` from the standard library.

<p align="center">
  <a href="https://github.com/omkarjaliparthi/kriya-go/actions/workflows/ci.yml"><img src="https://github.com/omkarjaliparthi/kriya-go/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://pkg.go.dev/github.com/omkarjaliparthi/kriya-go"><img src="https://img.shields.io/badge/pkg.go.dev-docs-007d9c?style=flat-square" /></a>
  <a href="https://kriya.insightsbyomkar.com/docs/api"><img src="https://img.shields.io/badge/API_docs-Scalar-6E56CF?style=flat-square" /></a>
  <a href="https://kriya.insightsbyomkar.com/pricing"><img src="https://img.shields.io/badge/Pricing-tiered-success?style=flat-square" /></a>
  <img src="https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square" />
  <img src="https://img.shields.io/badge/License-MIT-lightgrey?style=flat-square" />
  <img src="https://img.shields.io/badge/Dependencies-zero-brightgreen?style=flat-square" />
</p>

---

## Install

```bash
go get github.com/omkarjaliparthi/kriya-go@latest
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

    "github.com/omkarjaliparthi/kriya-go"
)

func main() {
    client := kriya.New("https://kriya.insightsbyomkar.com",
        kriya.WithAPIKey("YOUR_JWT_HERE"),
    )

    chart, err := client.NatalChart(context.Background(), kriya.Person{
        Datetime:  "1990-06-15T12:00:00Z", // strict ISO-8601, no naive locals
        Latitude:  51.5,
        Longitude: 0,
    })
    if err != nil {
        var apiErr *kriya.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API %d %s: %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
            return
        }
        panic(err)
    }
    fmt.Printf("%+v\n", chart)
}
```

Errors surface as `*kriya.APIError` with `Status`, `Code`, `Message`, `Details` — safe to unwrap with `errors.As`.

---

## Authentication

The API uses stateless HS256 JWTs. Each key encodes its own per-minute + per-day quotas — no round-trip to a key-store, horizontally scalable out of the box. Pass your key once at client construction:

```go
client := kriya.New(baseURL, kriya.WithAPIKey(os.Getenv("KRIYA_API_KEY")))
```

See the [API docs](https://kriya.insightsbyomkar.com/docs/api) for key minting + tier quotas.

---

## Endpoint coverage

**109+ v1 endpoints** across these domains. Grew from a 43-endpoint v1 launch (2026-04-15) to current scope across nine semver versions of disciplined iteration.

| Domain | Examples |
|---|---|
| **Western core** | `/chart/natal` · `/chart/wheel` · `/chart/extended` · `/transits` · `/transits/scan` · `/synastry` · `/composite` · `/composite/davison` · `/progressions` · `/returns/solar` · `/returns/lunar` · `/relocation` · `/harmonic` · `/draconic` |
| **Vedic** | `/vedic/chart` · `/vedic/panchanga` · `/vedic/varshaphala` · `/vedic/sade-sati` · `/vedic/yogas` · `/vargas` · `/dashas` · `/ashtakavarga` · `/shadbala` · `/chara-karakas` |
| **Jaimini** | `/jaimini/arudhas` · `/jaimini/aspects` · `/jaimini/chara-dasha` · `/jaimini/narayana-dasha` |
| **KP** | `/kp-sublords` · `/kp/cuspal-interlinks` · `/kp/ruling-planets` · `/kp/significators` |
| **Hellenistic** | `/profections` · `/zodiacal-releasing` · `/lots` · `/sensitive-points` |
| **Points & geometry** | `/positions` · `/aspects` · `/aspect-events` · `/pattern-events` · `/midpoints` · `/antiscia` · `/declinations` · `/lunar-mansions` · `/degrees/analyze` |
| **Electional / events** | `/muhurta` · `/eclipses` · `/eclipses/prenatal` · `/voc-moon` · `/planetary-hours` · `/stations` · `/sign-ingresses` · `/heliacal` · `/parans` · `/sun-rise-set` |
| **Bodies** | `/asteroids` · `/asteroids/extended` · `/fixed-stars` · `/stars/list` · `/bodies/extended` · `/heliocentric` · `/uranian` |
| **LLM-grounded** | `/reading/natal` · `/reading/transit` · `/reading/vedic` · `/ask-chart` · `/compatibility/narrative` |
| **Developer / admin** | `/keys/me` · `/keys/rotate` · `/usage/me` · `/jobs` · `/jobs/[id]` · `/webhooks` · `/accuracy` · `/calibrate` · `/openapi.json` |

Full live endpoint manifest at [`/openapi.json`](https://kriya.insightsbyomkar.com/openapi.json) or the [interactive Scalar docs](https://kriya.insightsbyomkar.com/docs/api).

Inputs are typed (`Person`, `NatalChartOpts`, `AspectPoint`); responses currently decode as `map[string]any` to match the API's rich, endpoint-specific shapes. Typed response structs are on the roadmap as the OpenAPI 3.1 spec stabilizes.

---

## Examples

Runnable examples live in [`examples/`](./examples):

```bash
export KRIYA_API_KEY="eyJ..."
go run ./examples/natal       # natal chart with typed error handling
go run ./examples/transits    # transits from a natal chart
go run ./examples/positions   # stateless body positions
```

## Design principles

- **Zero runtime dependencies.** Embed this in anything — serverless, edge, CLI, bot — without dragging a graph.
- **Context-first.** Every method takes `context.Context`. Cancel anywhere in the tree.
- **Errors are data.** `*kriya.APIError` carries the HTTP status, the machine-readable code, and a human message.
- **Idiomatic options.** Functional options (`WithAPIKey`, `WithHTTPClient`) — swap out the transport for tracing, retries, or testing.
- **Strict input parsing on the server.** ISO-8601 datetimes required — no ambiguous local times. Pair with Go's `time.Time.Format(time.RFC3339)` and you're set.

---

## Related

- **[Case study](https://github.com/omkarjaliparthi/insights-astrology-api-case-study)** — architecture, decisions, accuracy, sources
- **[TypeScript SDK](https://www.npmjs.com/package/kriya-astrology)** — `npm install kriya-astrology`
- **[Python SDK](https://pypi.org/project/kriya-astrology/)** — `pip install kriya-astrology`
- **[API docs](https://kriya.insightsbyomkar.com/docs/api)** — interactive Scalar explorer
- **[Pricing](https://kriya.insightsbyomkar.com/pricing)** — Developer (free) · Studio ($49/mo) · Scale (custom)

---

## License

MIT — see [`LICENSE`](./LICENSE). Free to embed in commercial applications.
