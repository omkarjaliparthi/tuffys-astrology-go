// Example: compute a natal chart with typed error handling.
//
//	TUFFYS_API_KEY=eyJ... go run ./examples/natal
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/omkarjaliparthi/tuffys-astrology-go/tuffys"
)

func main() {
	apiKey := os.Getenv("TUFFYS_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "set TUFFYS_API_KEY — mint one at https://tuffys-ai-astrology.vercel.app/pricing")
		os.Exit(1)
	}

	baseURL := os.Getenv("TUFFYS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://tuffys-ai-astrology.vercel.app"
	}

	client := tuffys.New(baseURL, tuffys.WithAPIKey(apiKey))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chart, err := client.NatalChart(ctx, tuffys.Person{
		Datetime:  "1990-06-15T12:00:00Z",
		Latitude:  51.5074,
		Longitude: -0.1278,
	}, tuffys.NatalChartOpts{HouseSystem: "placidus"})

	if err != nil {
		var apiErr *tuffys.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API %d %s: %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "transport error:", err)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(chart, "", "  ")
	fmt.Println(string(pretty))
}
