package amapclient

type NearbySearchRequest struct {
	Keywords string
	Location string
	Radius   int
	SortRule string
	Limit    int
	Filter   NearbySearchFilter
}

type NearbySearchFilter struct {
	MinCost   float64
	MaxCost   float64
	MinRating float64
}

type NearbySearchResponse struct {
	Status   string `json:"status,omitempty"`
	Info     string `json:"info,omitempty"`
	Infocode string `json:"infocode,omitempty"`
	Count    string `json:"count,omitempty"`
	POIs     []POI  `json:"pois"`
}

type POI struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Distance string   `json:"distance,omitempty"`
	Address  string   `json:"address,omitempty"`
	Location string   `json:"location,omitempty"`
	Province string   `json:"pname,omitempty"`
	City     string   `json:"cityname,omitempty"`
	County   string   `json:"adname,omitempty"`
	Type     string   `json:"type,omitempty"`
	TypeCode string   `json:"typecode,omitempty"`
	Business Business `json:"business"`
	Photos   []Photo  `json:"photos"`
}

type Photo struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

type Business struct {
	Area         string `json:"business_area,omitempty"`
	OpenTime     string `json:"opentime_today,omitempty"`
	OpenTimeWeek string `json:"opentime_week,omitempty"`
	Tel          string `json:"tel,omitempty"`
	Rating       string `json:"rating,omitempty"`
	Cost         string `json:"cost,omitempty"`
	ParkingType  string `json:"parking_type,omitempty"`
}
