package amapclient

import "encoding/json"

type NearbySearchRequest struct {
	Keywords string
	Types    string
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

type SearchRequest struct {
	Keywords string
	Types    string
	Region   string
	Limit    int
	Filter   NearbySearchFilter
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

type WeatherRequest struct {
	City       string
	Extensions string
}

type WeatherResponse struct {
	Status    string     `json:"status,omitempty"`
	Count     string     `json:"count,omitempty"`
	Info      string     `json:"info,omitempty"`
	Infocode  string     `json:"infocode,omitempty"`
	Lives     []Live     `json:"lives,omitempty"`
	Forecasts []Forecast `json:"forecasts,omitempty"`
}

type Live struct {
	Province      string `json:"province,omitempty"`
	City          string `json:"city,omitempty"`
	Adcode        string `json:"adcode,omitempty"`
	Weather       string `json:"weather,omitempty"`
	Temperature   string `json:"temperature,omitempty"`
	Winddirection string `json:"winddirection,omitempty"`
	Windpower     string `json:"windpower,omitempty"`
	Humidity      string `json:"humidity,omitempty"`
	Reporttime    string `json:"reporttime,omitempty"`
}

type Forecast struct {
	City       string `json:"city,omitempty"`
	Adcode     string `json:"adcode,omitempty"`
	Province   string `json:"province,omitempty"`
	Reporttime string `json:"reporttime,omitempty"`
	Casts      []Cast `json:"casts,omitempty"`
}

type Cast struct {
	Date         string `json:"date,omitempty"`
	Week         string `json:"week,omitempty"`
	Dayweather   string `json:"dayweather,omitempty"`
	Nightweather string `json:"nightweather,omitempty"`
	Daytemp      string `json:"daytemp,omitempty"`
	Nighttemp    string `json:"nighttemp,omitempty"`
	Daywind      string `json:"daywind,omitempty"`
	Nightwind    string `json:"nightwind,omitempty"`
	Daypower     string `json:"daypower,omitempty"`
	Nightpower   string `json:"nightpower,omitempty"`
}

type InputTipsRequest struct {
	Keywords string
	Types    string
	Location string
	City     string
	DataType string
}

type InputTipsResponse struct {
	Status   string `json:"status,omitempty"`
	Info     string `json:"info,omitempty"`
	Infocode string `json:"infocode,omitempty"`
	Count    string `json:"count,omitempty"`
	Tips     []Tip  `json:"tips"`
}

type Tip struct {
	ID       string     `json:"id,omitempty"`
	Name     string     `json:"name,omitempty"`
	District FlexString `json:"district,omitempty"`
	Adcode   FlexString `json:"adcode,omitempty"`
	Location FlexString `json:"location,omitempty"`
	Address  FlexString `json:"address,omitempty"`
}

// FlexString is a string field that tolerates AMAP shape inconsistencies:
//   - empty array "[]" instead of "" (e.g. busline tips return address/location
//     this way),
//   - bare JSON number (e.g. bicycling/electrobike directions return
//     step_distance as a number while driving/walking return it as a string).
//
// JSON marshaling falls back to the underlying string.
type FlexString string

func (s *FlexString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*s = FlexString(v)
		return nil
	}
	if c := data[0]; c == '-' || (c >= '0' && c <= '9') {
		*s = FlexString(data)
		return nil
	}
	return nil
}
