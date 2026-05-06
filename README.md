# AMAP-client

Go client and CLI for the [AMAP (高德地图) Web API](https://lbs.amap.com/api/webservice/summary).

Wraps a focused subset of AMAP's REST endpoints: nearby/keyword place search, input tips, geocoding (single + batch), live and forecast weather, and route planning across all five transport modes (driving, walking, bicycling, electrobike, transit).

## Installation

CLI:

```sh
go install github.com/minggeorgelei/AMAP-client/cmd/amap@latest
```

Library:

```sh
go get github.com/minggeorgelei/AMAP-client@latest
```

You'll need an [AMAP API key](https://console.amap.com/dev/key/app). Set it via `AMAP_API_KEY` (preferred) or pass `--key` on every call.

## CLI

```sh
export AMAP_API_KEY=your-key-here

amap nearby     --location="121.473667,31.230525" --keywords=咖啡 --radius=1000
amap search     --keywords=博物馆 --region=上海
amap tips       --keywords=人民广场 --city=上海
amap weather    --city=上海 --extensions=all
amap directions driving --origin="虹口区花园路168弄" --destination="兴业太古汇" --waypoints="静安大悦城;上海火车站"
amap directions transit --origin="虹口区花园路168弄" --destination="兴业太古汇"
amap directions walking --origin="121.473,31.230" --destination="121.448,31.226"
```

Origin, destination, `--location`, and waypoints all accept either `lng,lat` coordinates or a free-form address — addresses are geocoded automatically (in a single batched call when there are several, to avoid AMAP's per-second concurrency limit).

`--json` emits the raw response; default output is a colorized human-readable summary. `amap --help` (or `amap <subcommand> --help`) lists every flag.

## Library

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    amap "github.com/minggeorgelei/AMAP-client"
)

func main() {
    client := amap.NewClient(amap.Options{Key: os.Getenv("AMAP_API_KEY")})

    resp, err := client.DirectionsDriving(context.Background(), amap.DrivingRequest{
        DirectionsRequest: amap.DirectionsRequest{
            Origin:      "121.473667,31.230525",
            Destination: "121.448,31.226",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, path := range resp.Route.Paths {
        fmt.Printf("%s meters, %s seconds\n", path.Distance, path.Cost.Duration)
    }
}
```

The `Client` exposes one method per endpoint: `NearbySearch`, `Search`, `InputTips`, `Weather`, `Geocode`, `GeocodeBatch`, and `DirectionsDriving` / `DirectionsWalking` / `DirectionsBicycling` / `DirectionsElectrobike` / `DirectionsTransit`. See the per-method godoc for request/response details.

## License

MIT — see [LICENSE](LICENSE).
