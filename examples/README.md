# Examples

Runnable Go examples for the Tuffy's Astrology API client.

Each example is self-contained — `go run ./examples/<name>` from the repo root.

| Example | What it shows |
|---|---|
| [`natal/`](./natal) | Quickstart · `NatalChart` with typed error handling |
| [`transits/`](./transits) | Relational endpoint · `Transits` of a natal chart to a given datetime |
| [`positions/`](./positions) | Stateless body positions — no birth data needed |

## Auth

All examples read the API key from `TUFFYS_API_KEY`:

```bash
export TUFFYS_API_KEY="eyJ..."  # HS256 JWT issued by the API
go run ./examples/natal
```

For the free Developer tier, [mint a key](https://tuffys-ai-astrology.vercel.app/pricing). No credit card.

## Base URL

Examples default to the hosted API at `https://tuffys-ai-astrology.vercel.app`. Override with `TUFFYS_BASE_URL` to point at a local or custom deployment.
