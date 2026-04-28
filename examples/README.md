# Examples

Runnable Go examples for the Kriya astrology API client (formerly Tuffys).

Each example is self-contained — `go run ./examples/<name>` from the repo root.

| Example | What it shows |
|---|---|
| [`natal/`](./natal) | Quickstart · `NatalChart` with typed error handling |
| [`transits/`](./transits) | Relational endpoint · `Transits` of a natal chart to a given datetime |
| [`positions/`](./positions) | Stateless body positions — no birth data needed |

## Auth

All examples read the API key from `KRIYA_API_KEY`:

```bash
export KRIYA_API_KEY="eyJ..."  # HS256 JWT issued by the API
go run ./examples/natal
```

For the free Developer tier, [mint a key](https://kriya.insightsbyomkar.com/pricing). No credit card.

## Base URL

Examples default to the hosted API at `https://kriya.insightsbyomkar.com`. Override with `KRIYA_BASE_URL` to point at a local or custom deployment.
