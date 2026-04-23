package amapclient

import (
	"context"
	"net/url"
	"strconv"
)

func (c *Client) Search(ctx context.Context, req SearchRequest) (NearbySearchResponse, error) {
	if c.Key == "" {
		return NearbySearchResponse{}, ErrMissingAPIKey
	}
	if req.Keywords == "" && req.Types == "" {
		return NearbySearchResponse{}, ValidationError{Field: "keywords", Message: "keywords or types is required"}
	}
	if req.Filter.MinCost < 0 {
		return NearbySearchResponse{}, ValidationError{Field: "min-cost", Message: "must be non-negative"}
	}
	if req.Filter.MaxCost < 0 {
		return NearbySearchResponse{}, ValidationError{Field: "max-cost", Message: "must be non-negative"}
	}
	if req.Filter.MaxCost > 0 && req.Filter.MinCost > req.Filter.MaxCost {
		return NearbySearchResponse{}, ValidationError{Field: "min-cost", Message: "must not exceed max-cost"}
	}
	if req.Filter.MinRating < 0 || req.Filter.MinRating > 5 {
		return NearbySearchResponse{}, ValidationError{Field: "min-rating", Message: "must be between 0 and 5"}
	}

	limit := effectiveSearchLimit(req.Limit)

	if !filterActive(req.Filter) {
		return c.fetchSearchPage(ctx, req, limit, 1)
	}

	var collected []POI
	var last NearbySearchResponse
	for page := 1; len(collected) < limit; page++ {
		pageSize := defaultSearchPageSize
		resp, err := c.fetchSearchPage(ctx, req, pageSize, page)
		if err != nil {
			return NearbySearchResponse{}, err
		}
		if len(resp.POIs) == 0 {
			last = resp
			break
		}
		for _, poi := range resp.POIs {
			if matchFilter(req.Filter, poi.Business) {
				collected = append(collected, poi)
				if len(collected) >= limit {
					break
				}
			}
		}
		last = resp
		if len(resp.POIs) < pageSize {
			break
		}
	}

	last.POIs = collected
	last.Count = strconv.Itoa(len(collected))
	return last, nil
}

func (c *Client) fetchSearchPage(ctx context.Context, req SearchRequest, pageSize, pageNum int) (NearbySearchResponse, error) {
	params := url.Values{}
	if req.Keywords != "" {
		params.Set("keywords", req.Keywords)
	}
	if req.Types != "" {
		params.Set("types", req.Types)
	}
	if req.Region != "" {
		params.Set("region", req.Region)
	}
	params.Set("page_size", strconv.Itoa(pageSize))
	if pageNum > 1 {
		params.Set("page_num", strconv.Itoa(pageNum))
	}
	params.Set("show_fields", showFieldsNearbySearch)

	var resp NearbySearchResponse
	if err := c.Get(ctx, "/v5/place/text", params, &resp); err != nil {
		return NearbySearchResponse{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return resp, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	return resp, nil
}

func effectiveSearchLimit(l int) int {
	if l <= 0 {
		return defaultSearchLimit
	}
	if l > maxSearchLimit {
		return maxSearchLimit
	}
	return l
}
