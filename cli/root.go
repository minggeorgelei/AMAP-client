package cli

import "time"

// Root defines the CLI command tree.
type Root struct {
	Global GlobalOptions `embed:""`
	Nearby NearbyCmd     `cmd:"" help:"Search nearby places by location."`
	Search SearchCmd     `cmd:"" help:"Search places by keyword or type."`
}

type GlobalOptions struct {
	Key     string        `help:"AMAP API key." env:"AMAP_API_KEY"`
	BaseUrl string        `help:"Places API base URL." env:"AMAP_BASE_URL" default:"https://restapi.amap.com/v5"`
	Timeout time.Duration `help:"HTTP timeout." default:"10s"`
	JSON    bool          `help:"Output JSON."`
	NoColor bool          `help:"Disable color output."`
	Verbose bool          `help:"Verbose logging."`
	Version VersionFlag   `name:"version" help:"Print version and exit."`
}

type NearbyCmd struct {
	Location  string  `help:"Center point as 'longitude,latitude'." required:""`
	Keywords  string  `help:"Search keywords."`
	Radius    int     `help:"Search radius in meters (0-50000)." default:"5000"`
	SortRule  string  `help:"Sort rule: distance or weight." default:"distance" enum:"distance,weight"`
	Limit     int     `help:"Number of POIs to return (max 20)."`
	MinCost   float64 `name:"min-cost" help:"Minimum per-person cost; excludes POIs without cost data."`
	MaxCost   float64 `name:"max-cost" help:"Maximum per-person cost; excludes POIs without cost data."`
	MinRating float64 `name:"min-rating" help:"Minimum rating 0-5; excludes POIs without rating data."`
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
