# Tuffy's Astrology — Go SDK

Official Go client for the [Insights by Omkar Astrology API](https://tuffys-ai-astrology.vercel.app/docs).
Zero-dependency — only `net/http` and `encoding/json` from the standard library.

## Install

Once published to your GitHub org:

```bash
go get github.com/omkarjaliparthi/tuffys-astrology-go/tuffys@latest
```

Until then, vendor this directory into your project. See `SETUP.md` for publishing steps.

## Usage

```go
package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/omkarjaliparthi/tuffys-astrology-go/tuffys"
)

func main() {
    client := tuffys.New("https://your-host.example.com",
        tuffys.WithAPIKey("optional-key"),
    )

    chart, err := client.NatalChart(context.Background(), tuffys.Person{
        Datetime:  "1990-06-15T12:00:00Z",
        Latitude:  51.5,
        Longitude: 0,
    })
    if err != nil {
        var apiErr *tuffys.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API %d %s: %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
        }
        panic(err)
    }
    fmt.Printf("%+v\n", chart)
}
```

Errors surface as `*tuffys.APIError` with `Status`, `Code`, `Message`, `Details`.

## Requires

Go 1.21+. No external dependencies.

## License

MIT — see `LICENSE`.
