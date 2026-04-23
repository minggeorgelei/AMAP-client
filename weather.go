package amapclient

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
)

const (
	WeatherExtensionsBase = "base"
	WeatherExtensionsAll  = "all"
)

var adcodePattern = regexp.MustCompile(`^\d{6}$`)

// Weather fetches live or forecast weather for a city. The City field accepts
// either a 6-digit adcode or a city/region name; names are resolved to an
// adcode via the AMAP geocoding API before the weather call.
func (c *Client) Weather(ctx context.Context, req WeatherRequest) (WeatherResponse, error) {
	if c.Key == "" {
		return WeatherResponse{}, ErrMissingAPIKey
	}
	if req.City == "" {
		return WeatherResponse{}, ValidationError{Field: "city", Message: "is required"}
	}
	ext := req.Extensions
	if ext == "" {
		ext = WeatherExtensionsBase
	}
	if ext != WeatherExtensionsBase && ext != WeatherExtensionsAll {
		return WeatherResponse{}, ValidationError{Field: "extensions", Message: "must be 'base' or 'all'"}
	}

	adcode := req.City
	if !adcodePattern.MatchString(adcode) {
		resolved, err := c.resolveAdcode(ctx, req.City)
		if err != nil {
			return WeatherResponse{}, err
		}
		adcode = resolved
	}

	params := url.Values{}
	params.Set("city", adcode)
	params.Set("extensions", ext)

	var resp WeatherResponse
	if err := c.Get(ctx, "/v3/weather/weatherInfo", params, &resp); err != nil {
		return WeatherResponse{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return resp, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	return resp, nil
}

func (c *Client) resolveAdcode(ctx context.Context, address string) (string, error) {
	params := url.Values{}
	params.Set("address", address)

	var resp geocodeResponse
	if err := c.Get(ctx, "/v3/geocode/geo", params, &resp); err != nil {
		return "", err
	}
	if resp.Status != "" && resp.Status != "1" {
		return "", APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	if len(resp.Geocodes) == 0 || resp.Geocodes[0].Adcode == "" {
		return "", ValidationError{Field: "city", Message: fmt.Sprintf("could not resolve %q to an adcode", address)}
	}
	return resp.Geocodes[0].Adcode, nil
}
