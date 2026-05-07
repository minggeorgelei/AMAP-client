package amapclient

import (
	"context"
	"maps"
	"net/url"
	"strconv"
)

// DirectionsRequest is the common payload for all five direction modes.
type DirectionsRequest struct {
	Origin      string
	Destination string
}

// Per-mode show_fields defaults. Driving gets tmcs (live traffic) on top of
// the common cost because it's the only mode where road conditions
// vary meaningfully along the route.
const (
	showFieldsDriving    = "cost,tmcs,navi"
	showFieldsDirections = "cost,navi"
)

type DrivingRequest struct {
	DirectionsRequest
	Strategy  string
	Waypoints string
	Plate     string
}

type WalkingRequest struct {
	DirectionsRequest
	AlternativeRoute int
}

type BicyclingRequest struct {
	DirectionsRequest
	AlternativeRoute int
}

type ElectrobikeRequest struct {
	DirectionsRequest
	AlternativeRoute int
}

type TransitRequest struct {
	DirectionsRequest
	City1            string
	City2            string
	Strategy         string
	AlternativeRoute int
}

type DirectionsResponse struct {
	Status   string `json:"status,omitempty"`
	Info     string `json:"info,omitempty"`
	Infocode string `json:"infocode,omitempty"`
	Count    string `json:"count,omitempty"`
	Route    Route  `json:"route"`
}

type Route struct {
	Origin      string    `json:"origin,omitempty"`
	Destination string    `json:"destination,omitempty"`
	TaxiCost    string    `json:"taxi_cost,omitempty"`
	Paths       []Path    `json:"paths,omitempty"`
	Transits    []Transit `json:"transits,omitempty"`
}

type Path struct {
	Distance    string    `json:"distance,omitempty"`
	Restriction string    `json:"restriction,omitempty"`
	Cost        *PathCost `json:"cost,omitempty"`
	Tmcs        []Tmc     `json:"tmcs,omitempty"`
	Steps       []Step    `json:"steps,omitempty"`
}

type Tmc struct {
	TmcStatus   string `json:"tmc_status,omitempty"`
	TmcDistance string `json:"tmc_distance,omitempty"`
	TmcPolyline string `json:"tmc_polyline,omitempty"`
}

type PathCost struct {
	Duration      string `json:"duration,omitempty"`
	Tolls         string `json:"tolls,omitempty"`
	TollDistance  string `json:"toll_distance,omitempty"`
	TollRoad      string `json:"toll_road,omitempty"`
	TrafficLights string `json:"traffic_lights,omitempty"`
	Taxi          string `json:"taxi,omitempty"`
}

type Step struct {
	Instruction  string     `json:"instruction,omitempty"`
	Orientation  string     `json:"orientation,omitempty"`
	RoadName     string     `json:"road_name,omitempty"`
	StepDistance FlexString `json:"step_distance,omitempty"`
	Cost         *StepCost  `json:"cost,omitempty"`
	Navi         *StepNavi  `json:"navi,omitempty"`
	Polyline     string     `json:"polyline,omitempty"`
}

type StepNavi struct {
	Action          string `json:"action,omitempty"`
	AssistantAction string `json:"assistant_action,omitempty"`
}

type StepCost struct {
	Duration string `json:"duration,omitempty"`
}

type Transit struct {
	Distance        FlexString       `json:"distance,omitempty"`
	WalkingDistance FlexString       `json:"walking_distance,omitempty"`
	NightFlag       string           `json:"nightflag,omitempty"`
	Cost            *TransitCost     `json:"cost,omitempty"`
	Segments        []TransitSegment `json:"segments,omitempty"`
}

type TransitCost struct {
	Duration   FlexString `json:"duration,omitempty"`
	TaxiFee    FlexString `json:"taxi_fee,omitempty"`
	TransitFee FlexString `json:"transit_fee,omitempty"`
}

// TransitSegment is a single leg in a public-transit plan. Multiple sub-blocks
// (walking + bus, etc.) can be populated together — e.g. a leg that walks to a
// stop and then rides the bus.
type TransitSegment struct {
	Walking *TransitWalking `json:"walking,omitempty"`
	Bus     *TransitBus     `json:"bus,omitempty"`
	Railway *TransitRailway `json:"railway,omitempty"`
	Taxi    *TransitTaxi    `json:"taxi,omitempty"`
}

type TransitWalking struct {
	Distance FlexString `json:"distance,omitempty"`
	Cost     *StepCost  `json:"cost,omitempty"`
}

// TransitBus carries one or more candidate bus lines for the same leg — the
// caller picks one (e.g. "take line 1 OR line 4" for the same hop).
type TransitBus struct {
	Buslines []Busline `json:"buslines,omitempty"`
}

type Busline struct {
	Name          string      `json:"name,omitempty"`
	Type          string      `json:"type,omitempty"`
	Distance      FlexString  `json:"distance,omitempty"`
	Cost          *StepCost   `json:"cost,omitempty"`
	DepartureStop TransitStop `json:"departure_stop"`
	ArrivalStop   TransitStop `json:"arrival_stop"`
}

type TransitStop struct {
	Name     string `json:"name,omitempty"`
	Location string `json:"location,omitempty"`
}

type TransitRailway struct {
	Name     string     `json:"name,omitempty"`
	Distance FlexString `json:"distance,omitempty"`
}

type TransitTaxi struct {
	Price     FlexString `json:"price,omitempty"`
	DriveTime FlexString `json:"drivetime,omitempty"`
	Distance  FlexString `json:"distance,omitempty"`
	StartName string     `json:"startname,omitempty"`
	EndName   string     `json:"endname,omitempty"`
}

func (c *Client) DirectionsDriving(ctx context.Context, req DrivingRequest) (DirectionsResponse, error) {
	extra := url.Values{}
	extra.Set("show_fields", showFieldsDriving)
	if req.Strategy != "" {
		extra.Set("strategy", req.Strategy)
	}
	if req.Waypoints != "" {
		extra.Set("waypoints", req.Waypoints)
	}
	if req.Plate != "" {
		extra.Set("plate", req.Plate)
	}
	return c.directions(ctx, "/v5/direction/driving", extra, req.DirectionsRequest)
}

func (c *Client) DirectionsWalking(ctx context.Context, req WalkingRequest) (DirectionsResponse, error) {
	extra := url.Values{}
	extra.Set("show_fields", showFieldsDirections)
	if err := setAlternativeRoute(extra, "alternative_route", req.AlternativeRoute, 3); err != nil {
		return DirectionsResponse{}, err
	}
	return c.directions(ctx, "/v5/direction/walking", extra, req.DirectionsRequest)
}

func (c *Client) DirectionsBicycling(ctx context.Context, req BicyclingRequest) (DirectionsResponse, error) {
	extra := url.Values{}
	extra.Set("show_fields", showFieldsDirections)
	if err := setAlternativeRoute(extra, "alternative_route", req.AlternativeRoute, 3); err != nil {
		return DirectionsResponse{}, err
	}
	return c.directions(ctx, "/v5/direction/bicycling", extra, req.DirectionsRequest)
}

func (c *Client) DirectionsElectrobike(ctx context.Context, req ElectrobikeRequest) (DirectionsResponse, error) {
	extra := url.Values{}
	extra.Set("show_fields", showFieldsDirections)
	if err := setAlternativeRoute(extra, "alternative_route", req.AlternativeRoute, 3); err != nil {
		return DirectionsResponse{}, err
	}
	return c.directions(ctx, "/v5/direction/electrobike", extra, req.DirectionsRequest)
}

func (c *Client) DirectionsTransit(ctx context.Context, req TransitRequest) (DirectionsResponse, error) {
	if req.City1 == "" {
		return DirectionsResponse{}, ValidationError{Field: "city1", Message: "is required"}
	}
	if req.City2 == "" {
		return DirectionsResponse{}, ValidationError{Field: "city2", Message: "is required"}
	}
	extra := url.Values{}
	extra.Set("show_fields", showFieldsDirections)
	extra.Set("city1", req.City1)
	extra.Set("city2", req.City2)
	if req.Strategy != "" {
		extra.Set("strategy", req.Strategy)
	}
	// Transit's documented param name is the camel-case "AlternativeRoute" and
	// allows 1-10 (default 5), unlike the snake-case 1-3 used by the other modes.
	if err := setAlternativeRoute(extra, "AlternativeRoute", req.AlternativeRoute, 10); err != nil {
		return DirectionsResponse{}, err
	}
	return c.directions(ctx, "/v5/direction/transit/integrated", extra, req.DirectionsRequest)
}

func (c *Client) directions(ctx context.Context, path string, extra url.Values, base DirectionsRequest) (DirectionsResponse, error) {
	if c.Key == "" {
		return DirectionsResponse{}, ErrMissingAPIKey
	}
	if base.Origin == "" {
		return DirectionsResponse{}, ValidationError{Field: "origin", Message: "is required"}
	}
	if base.Destination == "" {
		return DirectionsResponse{}, ValidationError{Field: "destination", Message: "is required"}
	}

	params := url.Values{}
	params.Set("origin", base.Origin)
	params.Set("destination", base.Destination)
	maps.Copy(params, extra)

	var resp DirectionsResponse
	if err := c.Get(ctx, path, params, &resp); err != nil {
		return DirectionsResponse{}, err
	}
	if resp.Status != "" && resp.Status != "1" {
		return resp, APIError{InfoCode: resp.Infocode, Info: resp.Info}
	}
	return resp, nil
}

func setAlternativeRoute(params url.Values, key string, value, max int) error {
	if value == 0 {
		return nil
	}
	if value < 1 || value > max {
		return ValidationError{Field: "alternative-route", Message: "must be between 1 and " + strconv.Itoa(max)}
	}
	params.Set(key, strconv.Itoa(value))
	return nil
}
