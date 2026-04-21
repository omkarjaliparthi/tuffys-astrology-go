// Example: geocentric ecliptic positions of selected bodies at a given UTC instant.
//
//	TUFFYS_API_KEY=eyJ... go run ./examples/positions
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
		fmt.Fprintln(os.Stderr, "set TUFFYS_API_KEY")
		os.Exit(1)
	}

	baseURL := os.Getenv("TUFFYS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://tuffys-ai-astrology.vercel.app"
	}

	client := tuffys.New(baseURL, tuffys.WithAPIKey(apiKey))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now().UTC().Format(time.RFC3339)

	// Omit the bodies slice to get all ten.
	result, err := client.Positions(ctx, now, "sun", "moon", "mercury", "venus", "mars")
	if err != nil {
		var apiErr *tuffys.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API %d %s: %s\n", apiErr.Status, apiErr.Code, apiErr.Message)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "transport error:", err)
		os.Exit(1)
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(pretty))
}
