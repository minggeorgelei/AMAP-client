package cli

import (
	"fmt"
	"strconv"
	"strings"

	amapclient "github.com/minggeorgelei/AMAP-client"
)

func renderNearby(color Color, response amapclient.NearbySearchResponse) string {
	if len(response.POIs) == 0 {
		return "No nearby places found.\n"
	}

	var b strings.Builder
	for i, poi := range response.POIs {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(color.Bold(fmt.Sprintf("#%d ", i+1)))
		b.WriteString(color.Cyan(poi.Name))
		if poi.Distance != "" {
			b.WriteString("  ")
			b.WriteString(color.Dim("(" + formatDistance(poi.Distance) + ")"))
		}
		b.WriteString("\n")

		region := strings.Join(nonEmpty(poi.Province, poi.City, poi.County), " ")
		if poi.Business.Area != "" {
			if region != "" {
				region += " / " + poi.Business.Area
			} else {
				region = poi.Business.Area
			}
		}
		if region != "" {
			writeField(&b, color, "Region ", region)
		}
		if poi.Address != "" {
			writeField(&b, color, "Address", poi.Address)
		}
		if poi.Type != "" {
			writeField(&b, color, "Type   ", poi.Type)
		}
		if poi.Location != "" {
			writeField(&b, color, "Coord  ", poi.Location)
		}

		var meta []string
		if poi.Business.Rating != "" {
			meta = append(meta, color.Yellow("* "+poi.Business.Rating))
		}
		if poi.Business.Cost != "" {
			meta = append(meta, color.Green("$"+poi.Business.Cost))
		}
		if len(meta) > 0 {
			b.WriteString("  ")
			b.WriteString(color.Dim("Rating : "))
			b.WriteString(strings.Join(meta, "  "))
			b.WriteString("\n")
		}
		if poi.Business.Tel != "" {
			writeField(&b, color, "Tel    ", poi.Business.Tel)
		}
		if poi.Business.OpenTimeWeek != "" {
			writeField(&b, color, "Hours  ", poi.Business.OpenTimeWeek)
		}
		if poi.Business.ParkingType != "" {
			writeField(&b, color, "Parking", poi.Business.ParkingType)
		}
		for _, photo := range poi.Photos {
			value := photo.URL
			if photo.Title != "" {
				value = fmt.Sprintf("%s  %s", photo.URL, color.Dim("("+photo.Title+")"))
			}
			writeField(&b, color, "Photo  ", value)
		}
	}
	return b.String()
}

func renderTips(color Color, response amapclient.InputTipsResponse) string {
	if len(response.Tips) == 0 {
		return "No tips found.\n"
	}

	var b strings.Builder
	for i, tip := range response.Tips {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(color.Bold(fmt.Sprintf("#%d", i+1)))
		b.WriteString("\n")
		if tip.ID != "" {
			writeField(&b, color, "ID     ", tip.ID)
		}
		if tip.Name != "" {
			writeField(&b, color, "Name   ", color.Cyan(tip.Name))
		}
		if tip.District != "" {
			writeField(&b, color, "Region ", string(tip.District))
		}
		if tip.Adcode != "" {
			writeField(&b, color, "Adcode ", string(tip.Adcode))
		}
		if tip.Location != "" {
			writeField(&b, color, "Coord  ", string(tip.Location))
		}
		if tip.Address != "" {
			writeField(&b, color, "Address", string(tip.Address))
		}
	}
	return b.String()
}

func renderWeather(color Color, response amapclient.WeatherResponse) string {
	if len(response.Lives) == 0 && len(response.Forecasts) == 0 {
		return "No weather data.\n"
	}

	var b strings.Builder
	for i, live := range response.Lives {
		if i > 0 {
			b.WriteString("\n")
		}
		header := strings.Join(nonEmpty(live.Province, live.City), " ")
		if live.Adcode != "" {
			header += " " + color.Dim("("+live.Adcode+")")
		}
		b.WriteString(color.Bold("Live  "))
		b.WriteString(color.Cyan(header))
		b.WriteString("\n")
		if live.Weather != "" || live.Temperature != "" {
			wx := live.Weather
			if live.Temperature != "" {
				if wx != "" {
					wx += "  "
				}
				wx += color.Yellow(live.Temperature + "°C")
			}
			writeField(&b, color, "Cond   ", wx)
		}
		if live.Winddirection != "" || live.Windpower != "" {
			writeField(&b, color, "Wind   ", strings.TrimSpace(live.Winddirection+"风 "+live.Windpower+"级"))
		}
		if live.Humidity != "" {
			writeField(&b, color, "Humid  ", live.Humidity+"%")
		}
		if live.Reporttime != "" {
			writeField(&b, color, "Report ", live.Reporttime)
		}
	}

	for i, fc := range response.Forecasts {
		if i > 0 || len(response.Lives) > 0 {
			b.WriteString("\n")
		}
		header := strings.Join(nonEmpty(fc.Province, fc.City), " ")
		if fc.Adcode != "" {
			header += " " + color.Dim("("+fc.Adcode+")")
		}
		b.WriteString(color.Bold("Forecast  "))
		b.WriteString(color.Cyan(header))
		b.WriteString("\n")
		if fc.Reporttime != "" {
			writeField(&b, color, "Report ", fc.Reporttime)
		}
		for _, cast := range fc.Casts {
			day := strings.TrimSpace(cast.Date + " (星期" + convertWeekDay(cast.Week) + ")")
			b.WriteString("  ")
			b.WriteString(color.Dim(day + ": "))
			b.WriteString(fmt.Sprintf("%s %s°C / %s %s°C",
				cast.Dayweather, cast.Daytemp, cast.Nightweather, cast.Nighttemp))
			wind := strings.TrimSpace(cast.Daywind + "风 " + cast.Daypower + "级 / " + cast.Nightwind + "风 " + cast.Nightpower + "级")
			if wind != "" {
				b.WriteString("  ")
				b.WriteString(color.Dim(wind))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func renderDirections(color Color, response amapclient.DirectionsResponse) string {
	route := response.Route
	if len(route.Paths) == 0 && len(route.Transits) == 0 {
		return "No route found.\n"
	}

	var b strings.Builder
	header := strings.Join(nonEmpty(route.Origin, route.Destination), " → ")
	if header != "" {
		b.WriteString(color.Bold("Route  "))
		b.WriteString(color.Cyan(header))
		b.WriteString("\n")
	}
	if route.TaxiCost != "" {
		writeField(&b, color, "Taxi   ", color.Green("$"+route.TaxiCost))
	}

	for i, path := range route.Paths {
		b.WriteString("\n")
		b.WriteString(color.Bold(fmt.Sprintf("Path #%d", i+1)))
		b.WriteString("\n")
		if path.Distance != "" {
			writeField(&b, color, "Dist   ", formatDistance(path.Distance))
		}
		if path.Cost != nil {
			if path.Cost.Duration != "" {
				writeField(&b, color, "Time   ", color.Yellow(formatDuration(path.Cost.Duration)))
			}
			if path.Cost.Tolls != "" {
				writeField(&b, color, "Tolls  ", color.Green("$"+path.Cost.Tolls))
			}
			if path.Cost.TollDistance != "" {
				writeField(&b, color, "TollKm ", formatDistance(path.Cost.TollDistance))
			}
			if path.Cost.TollRoad != "" {
				writeField(&b, color, "TollRd ", path.Cost.TollRoad)
			}
			if path.Cost.TrafficLights != "" {
				writeField(&b, color, "Lights ", path.Cost.TrafficLights)
			}
			if path.Cost.Taxi != "" {
				writeField(&b, color, "Taxi   ", color.Green("$"+path.Cost.Taxi))
			}
		}
		if path.Restriction != "" && path.Restriction != "0" {
			writeField(&b, color, "Limit  ", color.Yellow("restricted"))
		}
		writeSteps(&b, color, path.Steps)
	}

	for i, transit := range route.Transits {
		b.WriteString("\n")
		b.WriteString(color.Bold(fmt.Sprintf("Transit #%d", i+1)))
		b.WriteString("\n")
		if transit.Distance != "" {
			writeField(&b, color, "Dist   ", formatDistance(string(transit.Distance)))
		}
		if transit.WalkingDistance != "" {
			writeField(&b, color, "Walk   ", formatDistance(string(transit.WalkingDistance)))
		}
		if transit.NightFlag == "1" {
			writeField(&b, color, "Night  ", color.Yellow("yes"))
		}
		if transit.Cost != nil {
			if transit.Cost.Duration != "" {
				writeField(&b, color, "Time   ", color.Yellow(formatDuration(string(transit.Cost.Duration))))
			}
			if transit.Cost.TransitFee != "" {
				writeField(&b, color, "Fare   ", color.Green("$"+string(transit.Cost.TransitFee)))
			}
			if transit.Cost.TaxiFee != "" {
				writeField(&b, color, "Taxi   ", color.Green("$"+string(transit.Cost.TaxiFee)))
			}
		}
		for j, seg := range transit.Segments {
			writeTransitSegment(&b, color, j+1, seg)
		}
	}
	return b.String()
}

func writeTransitSegment(b *strings.Builder, color Color, index int, seg amapclient.TransitSegment) {
	prefix := fmt.Sprintf("  %d. ", index)
	if seg.Walking != nil && seg.Walking.Distance != "" {
		b.WriteString(prefix)
		b.WriteString(color.Dim("walk   "))
		b.WriteString(formatDistance(string(seg.Walking.Distance)))
		if seg.Walking.Cost != nil && seg.Walking.Cost.Duration != "" {
			b.WriteString(color.Dim(" (" + formatDuration(string(seg.Walking.Cost.Duration)) + ")"))
		}
		b.WriteString("\n")
		prefix = "     "
	}
	if seg.Bus != nil && len(seg.Bus.Buslines) > 0 {
		bl := seg.Bus.Buslines[0]
		b.WriteString(prefix)
		b.WriteString(color.Dim(buslineLabel(bl.Name)))
		b.WriteString(color.Cyan(bl.Name))
		if bl.Distance != "" {
			b.WriteString(color.Dim(" (" + formatDistance(string(bl.Distance))))
			if bl.Cost != nil && bl.Cost.Duration != "" {
				b.WriteString(color.Dim(", " + formatDuration(bl.Cost.Duration)))
			}
			b.WriteString(color.Dim(")"))
		}
		b.WriteString("\n")
		if bl.DepartureStop.Name != "" || bl.ArrivalStop.Name != "" {
			b.WriteString("        ")
			b.WriteString(color.Dim(bl.DepartureStop.Name + " → " + bl.ArrivalStop.Name))
			b.WriteString("\n")
		}
		if len(seg.Bus.Buslines) > 1 {
			alts := make([]string, 0, len(seg.Bus.Buslines)-1)
			for _, b := range seg.Bus.Buslines[1:] {
				alts = append(alts, b.Name)
			}
			b.WriteString("        ")
			b.WriteString(color.Dim("or: " + strings.Join(alts, ", ")))
			b.WriteString("\n")
		}
		prefix = "     "
	}
	if seg.Railway != nil && (seg.Railway.Name != "" || seg.Railway.Distance != "") {
		b.WriteString(prefix)
		b.WriteString(color.Dim("rail  "))
		b.WriteString(color.Cyan(seg.Railway.Name))
		if seg.Railway.Distance != "" {
			b.WriteString(color.Dim(" (" + string(seg.Railway.Distance) + "m)"))
		}
		b.WriteString("\n")
		prefix = "     "
	}
	if seg.Taxi != nil && seg.Taxi.Distance != "" {
		b.WriteString(prefix)
		b.WriteString(color.Dim("taxi  "))
		b.WriteString(string(seg.Taxi.Distance) + "m")
		if seg.Taxi.Price != "" {
			b.WriteString("  " + color.Green("$"+string(seg.Taxi.Price)))
		}
		b.WriteString("\n")
	}
}

func buslineLabel(transitType string) string {
	if strings.Contains(transitType, "地铁") {
		return "metro  "
	}
	return "bus    "
}

func writeSteps(b *strings.Builder, color Color, steps []amapclient.Step) {
	if len(steps) == 0 {
		return
	}
	b.WriteString("  ")
	b.WriteString(color.Dim("Steps  :"))
	b.WriteString("\n")
	for i, step := range steps {
		b.WriteString("    ")
		b.WriteString(color.Dim(fmt.Sprintf("%d.", i+1)))
		b.WriteString(" ")
		if step.Instruction != "" {
			b.WriteString(step.Instruction)
		} else if step.RoadName != "" {
			b.WriteString(step.RoadName)
		}
		var meta []string
		if step.StepDistance != "" {
			meta = append(meta, string(step.StepDistance)+"m")
		}
		if step.Cost != nil && step.Cost.Duration != "" {
			meta = append(meta, formatDuration(step.Cost.Duration))
		}
		if len(meta) > 0 {
			b.WriteString("  ")
			b.WriteString(color.Dim("(" + strings.Join(meta, ", ") + ")"))
		}
		b.WriteString("\n")
	}
}

func convertWeekDay(week string) string {
	switch week {
	case "1":
		return "一"
	case "2":
		return "二"
	case "3":
		return "三"
	case "4":
		return "四"
	case "5":
		return "五"
	case "6":
		return "六"
	case "7":
		return "日"
	default:
		return ""
	}
}

func writeField(b *strings.Builder, color Color, label, value string) {
	b.WriteString("  ")
	b.WriteString(color.Dim(label + ": "))
	b.WriteString(value)
	b.WriteString("\n")
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func formatDistance(s string) string {
	if s == "" {
		return ""
	}
	meters, err := strconv.ParseInt(s, 10, 64)
	if err != nil || meters < 0 {
		return s + "m"
	}
	if meters < 1000 {
		return fmt.Sprintf("%dm", meters)
	}
	return fmt.Sprintf("%.3fkm", float64(meters)/1000)
}

func formatDuration(s string) string {
	if s == "" {
		return ""
	}
	secs, err := strconv.ParseInt(s, 10, 64)
	if err != nil || secs < 0 {
		return s + "s"
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	sec := secs % 60
	var buf strings.Builder
	if h > 0 {
		fmt.Fprintf(&buf, "%dh", h)
	}
	if m > 0 {
		fmt.Fprintf(&buf, "%dm", m)
	}
	if sec > 0 {
		fmt.Fprintf(&buf, "%ds", sec)
	}
	return buf.String()
}
