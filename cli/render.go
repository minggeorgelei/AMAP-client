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
