package amapclient

import (
	"context"
	"net/url"
)

// InputTips queries AMAP's /v3/assistant/inputtips endpoint for suggestion
// hints matching a search keyword. Location is optional and only biases
// results when City is also set.
func (c *Client) InputTips(ctx context.Context, req InputTipsRequest) (InputTipsResponse, error) {
	if c.Key == "" {
		return InputTipsResponse{}, ErrMissingAPIKey
	}
	if req.Keywords == "" {
		return InputTipsResponse{}, ValidationError{Field: "keywords", Message: "is required"}
	}

	params := url.Values{}
	params.Set("keywords", req.Keywords)
	if req.Types != "" {
		params.Set("type", req.Types)
	}
	if req.Location != "" {
		params.Set("location", req.Location)
	}
	if req.City != "" {
		params.Set("city", req.City)
	}
	if req.DataType != "" {
		params.Set("datatype", req.DataType)
	}

	var resp InputTipsResponse
	if err := c.Get(ctx, "/v3/assistant/inputtips", params, &resp); err != nil {
		return InputTipsResponse{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return resp, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	return resp, nil
}
