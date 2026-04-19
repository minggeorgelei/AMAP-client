package amapclient

import (
	"context"
	"net/url"
	"strconv"
)

const showFieldsNearbySearch = "business,photos"

func (c *Client) NearbySearch(ctx context.Context, req NearbySearchRequest) (NearbySearchResponse, error) {
	if c.Key == "" {
		return NearbySearchResponse{}, ErrMissingAPIKey
	}
	if req.Location == "" {
		return NearbySearchResponse{}, ValidationError{Field: "location", Message: "is required"}
	}
	if req.Radius < 0 || req.Radius > 50000 {
		return NearbySearchResponse{}, ValidationError{Field: "radius", Message: "must be between 0 and 50000"}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultNearbySearchLimit
	}
	if limit > maxNearbySearchLimit {
		limit = maxNearbySearchLimit
	}

	params := url.Values{}
	params.Set("location", req.Location)
	if req.Keywords != "" {
		params.Set("keywords", req.Keywords)
	}
	if req.Radius > 0 {
		params.Set("radius", strconv.Itoa(req.Radius))
	}
	if req.SortRule != "" {
		params.Set("sortrule", req.SortRule)
	}
	params.Set("page_size", strconv.Itoa(limit))
	params.Set("show_fields", showFieldsNearbySearch)

	var response NearbySearchResponse
	if err := c.Get(ctx, "/place/around", params, &response); err != nil {
		return NearbySearchResponse{}, err
	}
	if response.Status != "" && response.Status != "1" {
		return response, APIError{InfoCode: response.Infocode, Info: response.Info}
	}
	return response, nil
}
