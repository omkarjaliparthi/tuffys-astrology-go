// Package kriya is the Go client for Kriya — the Insights Astrology API by Insights by Omkar.
//
// Zero runtime dependencies beyond the standard library.
//
// Basic usage:
//
//	client := kriya.New("https://your-host", kriya.WithAPIKey("..."))
//	chart, err := client.NatalChart(ctx, kriya.Person{
//	    Datetime:  "1990-06-15T12:00:00Z",
//	    Latitude:  51.5,
//	    Longitude: 0,
//	})
package kriya

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to the Tuffy's Astrology HTTP API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the x-api-key header for authenticated requests.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithHTTPClient overrides the underlying http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// New constructs a Client pointing at baseURL (e.g. "https://kriya.example.com").
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// APIError represents a structured error returned by the API.
type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kriya API %d %s: %s", e.Status, e.Code, e.Message)
}

type apiErrorEnvelope struct {
	Error APIError `json:"error"`
}

// Person is a birth specification.
type Person struct {
	Datetime  string  `json:"datetime"`           // ISO 8601 with timezone
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AspectPoint struct {
	Key       string  `json:"key"`
	Longitude float64 `json:"longitude"`
}

type NatalChartOpts struct {
	HouseSystem string `json:"houseSystem,omitempty"` // placidus | porphyry | equal | whole-sign | koch
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var env apiErrorEnvelope
		if jerr := json.Unmarshal(raw, &env); jerr == nil && env.Error.Code != "" {
			env.Error.Status = resp.StatusCode
			return &env.Error
		}
		return errors.New(fmt.Sprintf("kriya API %d: %s", resp.StatusCode, string(raw)))
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// --------- Core ---------

// NatalChart returns a full natal chart (bodies, houses, aspects).
func (c *Client) NatalChart(ctx context.Context, person Person, opts ...NatalChartOpts) (map[string]any, error) {
	body := map[string]any{
		"datetime":  person.Datetime,
		"latitude":  person.Latitude,
		"longitude": person.Longitude,
	}
	if len(opts) > 0 && opts[0].HouseSystem != "" {
		body["houseSystem"] = opts[0].HouseSystem
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/chart/natal", body, &out)
	return out, err
}

// ExtendedChart adds nodes, Lilith, lots, vertex, midpoints, declinations.
func (c *Client) ExtendedChart(ctx context.Context, person Person, opts ...NatalChartOpts) (map[string]any, error) {
	body := map[string]any{
		"datetime":  person.Datetime,
		"latitude":  person.Latitude,
		"longitude": person.Longitude,
	}
	if len(opts) > 0 && opts[0].HouseSystem != "" {
		body["houseSystem"] = opts[0].HouseSystem
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/chart/extended", body, &out)
	return out, err
}

// Positions returns geocentric ecliptic positions; omit `bodies` to get all 10.
func (c *Client) Positions(ctx context.Context, datetime string, bodies ...string) (map[string]any, error) {
	body := map[string]any{"datetime": datetime}
	if len(bodies) > 0 {
		body["bodies"] = bodies
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/positions", body, &out)
	return out, err
}

// Houses returns cusps + angles only.
func (c *Client) Houses(ctx context.Context, person Person, system string) (map[string]any, error) {
	body := map[string]any{
		"datetime":  person.Datetime,
		"latitude":  person.Latitude,
		"longitude": person.Longitude,
	}
	if system != "" {
		body["system"] = system
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/houses", body, &out)
	return out, err
}

// Aspects detects aspects between arbitrary longitudes.
func (c *Client) Aspects(ctx context.Context, points []AspectPoint) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/aspects", map[string]any{"points": points}, &out)
	return out, err
}

// --------- Relational ---------

func (c *Client) Transits(ctx context.Context, natal Person, transitDatetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/transits", map[string]any{
		"natal":           natal,
		"transitDatetime": transitDatetime,
	}, &out)
	return out, err
}

func (c *Client) Synastry(ctx context.Context, a, b Person) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/synastry", map[string]any{"personA": a, "personB": b}, &out)
	return out, err
}

func (c *Client) Composite(ctx context.Context, a, b Person) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/composite", map[string]any{"personA": a, "personB": b}, &out)
	return out, err
}

func (c *Client) SolarReturn(ctx context.Context, natal Person, yearsAfter int) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/returns/solar", map[string]any{
		"natal":      natal,
		"yearsAfter": yearsAfter,
	}, &out)
	return out, err
}

func (c *Client) LunarReturn(ctx context.Context, natal Person, monthsAfter int) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/returns/lunar", map[string]any{
		"natal":        natal,
		"monthsAfter":  monthsAfter,
	}, &out)
	return out, err
}

// --------- Vedic ---------

func (c *Client) VedicChart(ctx context.Context, person Person, ayanamsa string) (map[string]any, error) {
	body := map[string]any{
		"datetime":  person.Datetime,
		"latitude":  person.Latitude,
		"longitude": person.Longitude,
	}
	if ayanamsa != "" {
		body["ayanamsa"] = ayanamsa
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/vedic/chart", body, &out)
	return out, err
}

func (c *Client) Panchanga(ctx context.Context, datetime string, ayanamsa string) (map[string]any, error) {
	body := map[string]any{"datetime": datetime}
	if ayanamsa != "" {
		body["ayanamsa"] = ayanamsa
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/vedic/panchanga", body, &out)
	return out, err
}

func (c *Client) Muhurta(ctx context.Context, datetime, ayanamsa string) (map[string]any, error) {
	body := map[string]any{"datetime": datetime}
	if ayanamsa != "" {
		body["ayanamsa"] = ayanamsa
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/muhurta", body, &out)
	return out, err
}

func (c *Client) Dashas(ctx context.Context, person Person, system string) (map[string]any, error) {
	body := map[string]any{
		"datetime":  person.Datetime,
		"latitude":  person.Latitude,
		"longitude": person.Longitude,
	}
	if system != "" {
		body["system"] = system
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/dashas", body, &out)
	return out, err
}

// --------- Points & stars ---------

func (c *Client) TrueNode(ctx context.Context, datetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/true-node", map[string]any{"datetime": datetime}, &out)
	return out, err
}

func (c *Client) Asteroids(ctx context.Context, datetime string, asteroids ...string) (map[string]any, error) {
	body := map[string]any{"datetime": datetime}
	if len(asteroids) > 0 {
		body["asteroids"] = asteroids
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/asteroids", body, &out)
	return out, err
}

func (c *Client) FixedStars(ctx context.Context, datetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/fixed-stars", map[string]any{"datetime": datetime}, &out)
	return out, err
}

// --------- Eclipses ---------

func (c *Client) Eclipses(ctx context.Context, startDatetime, endDatetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/eclipses", map[string]any{
		"startDatetime": startDatetime,
		"endDatetime":   endDatetime,
	}, &out)
	return out, err
}

func (c *Client) PrenatalEclipses(ctx context.Context, datetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/eclipses/prenatal", map[string]any{"datetime": datetime}, &out)
	return out, err
}

// --------- Electional / Daily ---------

func (c *Client) PlanetaryHours(ctx context.Context, datetime string, lat, lon float64) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/planetary-hours", map[string]any{
		"datetime":  datetime,
		"latitude":  lat,
		"longitude": lon,
	}, &out)
	return out, err
}

func (c *Client) VOCMoon(ctx context.Context, datetime string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/voc-moon", map[string]any{"datetime": datetime}, &out)
	return out, err
}

// --------- Readings ---------

func (c *Client) Daily(ctx context.Context, natal Person, atDatetime string) (map[string]any, error) {
	body := map[string]any{"natal": natal}
	if atDatetime != "" {
		body["atDatetime"] = atDatetime
	}
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/daily", body, &out)
	return out, err
}

func (c *Client) Compatibility(ctx context.Context, a, b Person) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "POST", "/api/v1/compatibility", map[string]any{"personA": a, "personB": b}, &out)
	return out, err
}

// --------- Discovery ---------

func (c *Client) OpenAPISpec(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, "GET", "/api/v1/openapi.json", nil, &out)
	return out, err
}
