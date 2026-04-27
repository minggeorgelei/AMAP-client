package cli

import "time"

// Root defines the CLI command tree.
type Root struct {
	Global  GlobalOptions `embed:""`
	Nearby  NearbyCmd     `cmd:"" help:"Search nearby places by location."`
	Search  SearchCmd     `cmd:"" help:"Search places by keyword or type."`
	Tips    TipsCmd       `cmd:"" help:"Query input suggestion tips by keyword."`
	Weather WeatherCmd    `cmd:"" help:"Query live or forecast weather by city."`
}

type GlobalOptions struct {
	Key     string        `help:"AMAP API key." env:"AMAP_API_KEY"`
	BaseUrl string        `help:"AMAP API base URL." env:"AMAP_BASE_URL" default:"https://restapi.amap.com"`
	Timeout time.Duration `help:"HTTP timeout." default:"10s"`
	JSON    bool          `help:"Output JSON."`
	NoColor bool          `help:"Disable color output."`
	Verbose bool          `help:"Verbose logging."`
	Version VersionFlag   `name:"version" help:"Print version and exit."`
}

type NearbyCmd struct {
	Location  string  `help:"Center as 'longitude,latitude' or an address/place name (geocoded if not coordinates)." required:""`
	Keywords  string  `help:"Search keywords."`
	Types     string  `help:"POI type codes, pipe-separated."`
	Radius    int     `help:"Search radius in meters (0-50000)." default:"5000"`
	SortRule  string  `help:"Sort rule: distance or weight." default:"distance" enum:"distance,weight"`
	Limit     int     `help:"Number of POIs to return (max 20)."`
	MinCost   float64 `name:"min-cost" help:"Minimum per-person cost; excludes POIs without cost data."`
	MaxCost   float64 `name:"max-cost" help:"Maximum per-person cost; excludes POIs without cost data."`
	MinRating float64 `name:"min-rating" help:"Minimum rating 0-5; excludes POIs without rating data."`
}

type TipsCmd struct {
	Keywords string `help:"Search keywords." required:""`
	Types    string `help:"POI type codes (or names), pipe-separated."`
	Location string `help:"Bias center as 'longitude,latitude' or an address (geocoded if not coordinates); only takes effect when --city is set."`
	City     string `help:"City name, citycode, or adcode to bias search."`
	DataType string `name:"datatype" help:"Data type filter: all, poi, bus, busline (pipe-separated)." default:"all"`
}

type SearchCmd struct {
	Keywords  string  `help:"Search keywords (required if --types is empty)."`
	Types     string  `help:"POI type codes, pipe-separated (required if --keywords is empty)."`
	Region    string  `help:"Region name, citycode, or adcode to bias search."`
	Limit     int     `help:"Number of POIs to return (max 20)."`
	MinCost   float64 `name:"min-cost" help:"Minimum per-person cost; excludes POIs without cost data."`
	MaxCost   float64 `name:"max-cost" help:"Maximum per-person cost; excludes POIs without cost data."`
	MinRating float64 `name:"min-rating" help:"Minimum rating 0-5; excludes POIs without rating data."`
}

type WeatherCmd struct {
	City       string `help:"City adcode (6 digits) or name; names are resolved via geocoding." required:""`
	Extensions string `help:"Weather type: base (live) or all (forecast)." default:"base" enum:"base,all"`
}
