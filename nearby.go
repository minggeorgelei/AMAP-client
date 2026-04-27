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

	limit := effectiveNearbyLimit(req.Limit)

	if !filterActive(req.Filter) {
		return c.fetchNearbyPage(ctx, req, limit, 1)
	}

	var collected []POI
	var last NearbySearchResponse
	for page := 1; len(collected) < limit; page++ {
		pageSize := defaultNearbySearchPageSize
		resp, err := c.fetchNearbyPage(ctx, req, pageSize, page)
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

func (c *Client) fetchNearbyPage(ctx context.Context, req NearbySearchRequest, pageSize, pageNum int) (NearbySearchResponse, error) {
	params := url.Values{}
	params.Set("location", req.Location)
	if req.Keywords != "" {
		params.Set("keywords", req.Keywords)
	}
	if req.Types != "" {
		params.Set("types", req.Types)
	}
	if req.Radius > 0 {
		params.Set("radius", strconv.Itoa(req.Radius))
	}
	if req.SortRule != "" {
		params.Set("sortrule", req.SortRule)
	}
	params.Set("page_size", strconv.Itoa(pageSize))
	if pageNum > 1 {
		params.Set("page_num", strconv.Itoa(pageNum))
	}
	params.Set("show_fields", showFieldsNearbySearch)

	var resp NearbySearchResponse
	if err := c.Get(ctx, "/v5/place/around", params, &resp); err != nil {
		return NearbySearchResponse{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return resp, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	return resp, nil
}

func effectiveNearbyLimit(l int) int {
	if l <= 0 {
		return defaultNearbySearchLimit
	}
	if l > maxNearbySearchLimit {
		return maxNearbySearchLimit
	}
	return l
}

func filterActive(f NearbySearchFilter) bool {
	return f.MinCost > 0 || f.MaxCost > 0 || f.MinRating > 0
}

func matchFilter(f NearbySearchFilter, b Business) bool {
	if f.MinCost > 0 || f.MaxCost > 0 {
		if b.Cost == "" {
			return false
		}
		cost, err := strconv.ParseFloat(b.Cost, 64)
		if err != nil {
			return false
		}
		if f.MinCost > 0 && cost < f.MinCost {
			return false
		}
		if f.MaxCost > 0 && cost > f.MaxCost {
			return false
		}
	}
	if f.MinRating > 0 {
		if b.Rating == "" {
			return false
		}
		rating, err := strconv.ParseFloat(b.Rating, 64)
		if err != nil {
			return false
		}
		if rating < f.MinRating {
			return false
		}
	}
	return true
}
