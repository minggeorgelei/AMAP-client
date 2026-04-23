package cli

import (
	"fmt"
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
			b.WriteString(color.Dim(fmt.Sprintf("(%sm)", poi.Distance)))
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
