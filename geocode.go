package amapclient

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// GeocodeBatchLimit is AMAP's documented per-request cap for batch geocoding.
const GeocodeBatchLimit = 10

type GeocodeResult struct {
	FormattedAddress string
	Adcode           string
	Location         string
	Level            string
}

type geocodeResponse struct {
	Status   string    `json:"status,omitempty"`
	Info     string    `json:"info,omitempty"`
	Infocode string    `json:"infocode,omitempty"`
	Geocodes []geocode `json:"geocodes,omitempty"`
}

// geocode mirrors AMAP's /v3/geocode/geo result. city/district/street/number
// are typed as any because AMAP returns them as a string when populated and as
// an empty array ("[]") when missing — decoding into string would fail in the
// latter. Province/country/adcode/location/level are always strings.
type geocode struct {
	FormattedAddress string `json:"formatted_address,omitempty"`
	Adcode           string `json:"adcode,omitempty"`
	Location         string `json:"location,omitempty"`
	Level            string `json:"level,omitempty"`
}

// Geocode resolves a free-form address to coordinates and region metadata via
// AMAP's /v3/geocode/geo endpoint. The Level field indicates the granularity
// AMAP matched, so callers can warn when input
// was too vague.
func (c *Client) Geocode(ctx context.Context, address string) (GeocodeResult, error) {
	if c.Key == "" {
		return GeocodeResult{}, ErrMissingAPIKey
	}
	if address == "" {
		return GeocodeResult{}, ValidationError{Field: "address", Message: "is required"}
	}

	params := url.Values{}
	params.Set("address", address)

	var resp geocodeResponse
	if err := c.Get(ctx, "/v3/geocode/geo", params, &resp); err != nil {
		return GeocodeResult{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return GeocodeResult{}, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	if len(resp.Geocodes) == 0 {
		return GeocodeResult{}, ValidationError{Field: "address", Message: fmt.Sprintf("no geocode result for %q", address)}
	}
	g := resp.Geocodes[0]
	return GeocodeResult{
		FormattedAddress: g.FormattedAddress,
		Location:         g.Location,
		Level:            g.Level,
		Adcode:           g.Adcode,
	}, nil
}

// GeocodeBatch resolves up to GeocodeBatchLimit addresses in a single API call
// (address=A|B|C with batch=true). Results are returned in input order; an
// entry with an empty Location means AMAP couldn't match that address. Batching
// matters because callers running address-heavy flows (e.g. driving directions
// with several waypoints) otherwise risk tripping AMAP's per-second concurrency
// limit by firing many single-address calls back-to-back.
func (c *Client) GeocodeBatch(ctx context.Context, addresses []string) ([]GeocodeResult, error) {
	if c.Key == "" {
		return nil, ErrMissingAPIKey
	}
	if len(addresses) == 0 {
		return nil, ValidationError{Field: "addresses", Message: "is required"}
	}
	if len(addresses) > GeocodeBatchLimit {
		return nil, ValidationError{Field: "addresses", Message: fmt.Sprintf("supports at most %d addresses per call", GeocodeBatchLimit)}
	}
	for i, a := range addresses {
		if a == "" {
			return nil, ValidationError{Field: "addresses", Message: fmt.Sprintf("entry %d is empty", i)}
		}
		if strings.Contains(a, "|") {
			return nil, ValidationError{Field: "addresses", Message: fmt.Sprintf("entry %d contains a '|' which is the batch separator", i)}
		}
	}

	params := url.Values{}
	params.Set("address", strings.Join(addresses, "|"))
	params.Set("batch", "true")

	var resp geocodeResponse
	if err := c.Get(ctx, "/v3/geocode/geo", params, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return nil, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}

	results := make([]GeocodeResult, len(resp.Geocodes))
	for i, g := range resp.Geocodes {
		results[i] = GeocodeResult{
			FormattedAddress: g.FormattedAddress,
			Adcode:           g.Adcode,
			Location:         g.Location,
			Level:            g.Level,
		}
	}
	return results, nil
}
