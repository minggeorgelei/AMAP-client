package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/alecthomas/kong"
	amapclient "github.com/minggeorgelei/AMAP-client"
)

var coordPattern = regexp.MustCompile(`^-?\d+(\.\d+)?,-?\d+(\.\d+)?$`)

type App struct {
	client *amapclient.Client
	out    io.Writer
	error  io.Writer
	json   bool
	color  Color
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	root := Root{}
	exitCode := 0
	parser, err := kong.New(
		&root,
		kong.Name("AMAP-client"),
		kong.Description("Search and resolve places via the AMAP API."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true, Summary: true}),
		kong.Writers(stdout, stderr),
		kong.Exit(func(code int) {
			exitCode = code
			panic(exitSignal{code: code})
		}),
		kong.Vars{"version": Version},
	)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	ctx, exited, err := parseWithExit(parser, args, &exitCode)
	if exited {
		return exitCode
	}
	if err != nil {
		if parseErr, ok := err.(*kong.ParseError); ok {
			_ = parseErr.Context.PrintUsage(true)
			_, _ = fmt.Fprintln(stderr, parseErr.Error())
			return parseErr.ExitCode()
		}
		_, _ = fmt.Fprintln(stderr, err)
		return 2
	}
	if root.Global.JSON {
		// JSON output should never include ANSI escapes.
		root.Global.NoColor = true
	}

	client := amapclient.NewClient(amapclient.Options{
		Key:     root.Global.Key,
		BaseUrl: root.Global.BaseUrl,
		Timeout: root.Global.Timeout,
	})

	app := &App{
		client: client,
		out:    stdout,
		error:  stderr,
		json:   root.Global.JSON,
		color:  NewColor(colorEnabled(root.Global.NoColor)),
	}

	ctx.Bind(app)
	if err := ctx.Run(); err != nil {
		return handleError(stderr, err)
	}

	return 0
}

type exitSignal struct {
	code int
}

func handleError(writer io.Writer, err error) int {
	if err == nil {
		return 0
	}
	var validation amapclient.ValidationError
	if errors.As(err, &validation) {
		_, _ = fmt.Fprintln(writer, validation.Error())
		return 2
	}
	if errors.Is(err, amapclient.ErrMissingAPIKey) {
		_, _ = fmt.Fprintln(writer, err.Error())
		return 2
	}
	_, _ = fmt.Fprintln(writer, err.Error())
	return 1
}

func parseWithExit(parser *kong.Kong, args []string, exitCode *int) (ctx *kong.Context, exited bool, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if signal, ok := recovered.(exitSignal); ok {
				// kong uses Exit() hooks; convert to a normal return.
				if exitCode != nil {
					*exitCode = signal.code
				}
				exited = true
				ctx = nil
				err = nil
				return
			}
			panic(recovered)
		}
	}()
	ctx, err = parser.Parse(args)
	return ctx, exited, err
}

func (c *NearbyCmd) Run(app *App) error {
	ctx := context.Background()
	location, err := resolveLocation(ctx, app, c.Location)
	if err != nil {
		return err
	}
	req := amapclient.NearbySearchRequest{
		Location: location,
		Keywords: c.Keywords,
		Types:    c.Types,
		Radius:   c.Radius,
		SortRule: c.SortRule,
		Limit:    c.Limit,
		Filter: amapclient.NearbySearchFilter{
			MinCost:   c.MinCost,
			MaxCost:   c.MaxCost,
			MinRating: c.MinRating,
		},
	}
	response, err := app.client.NearbySearch(ctx, req)
	if err != nil {
		return err
	}
	return app.writePOIs(response)
}

func (c *TipsCmd) Run(app *App) error {
	ctx := context.Background()
	location, err := resolveLocation(ctx, app, c.Location)
	if err != nil {
		return err
	}
	req := amapclient.InputTipsRequest{
		Keywords: c.Keywords,
		Types:    c.Types,
		Location: location,
		City:     c.City,
		DataType: c.DataType,
	}
	response, err := app.client.InputTips(ctx, req)
	if err != nil {
		return err
	}
	return app.writeTips(response)
}

// resolveLocation accepts either an empty string, a "lng,lat" coordinate pair,
// or a free-form address/place name. Names are geocoded into coordinates.
func resolveLocation(ctx context.Context, app *App, loc string) (string, error) {
	if loc == "" || coordPattern.MatchString(loc) {
		return loc, nil
	}
	result, err := app.client.Geocode(ctx, loc)
	if err != nil {
		return "", fmt.Errorf("geocode %q: %w", loc, err)
	}
	if result.Location == "" {
		return "", fmt.Errorf("geocode %q: no coordinates returned", loc)
	}
	return result.Location, nil
}

func (c *SearchCmd) Run(app *App) error {
	req := amapclient.SearchRequest{
		Keywords: c.Keywords,
		Types:    c.Types,
		Region:   c.Region,
		Limit:    c.Limit,
		Filter: amapclient.NearbySearchFilter{
			MinCost:   c.MinCost,
			MaxCost:   c.MaxCost,
			MinRating: c.MinRating,
		},
	}
	response, err := app.client.Search(context.Background(), req)
	if err != nil {
		return err
	}
	return app.writePOIs(response)
}

func (c *DirectionsDrivingCmd) Run(app *App) error {
	ctx := context.Background()
	base, err := resolveDirectionsCommon(ctx, app, c.DirectionsCommon)
	if err != nil {
		return err
	}
	resp, err := app.client.DirectionsDriving(ctx, amapclient.DrivingRequest{
		DirectionsRequest: base,
		Strategy:          c.Strategy,
		Waypoints:         c.Waypoints,
		Plate:             c.Plate,
	})
	if err != nil {
		return err
	}
	return app.writeDirections(resp)
}

func (c *DirectionsWalkingCmd) Run(app *App) error {
	ctx := context.Background()
	base, err := resolveDirectionsCommon(ctx, app, c.DirectionsCommon)
	if err != nil {
		return err
	}
	resp, err := app.client.DirectionsWalking(ctx, amapclient.WalkingRequest{
		DirectionsRequest: base,
		AlternativeRoute:  c.AlternativeRoute,
	})
	if err != nil {
		return err
	}
	return app.writeDirections(resp)
}

func (c *DirectionsBicyclingCmd) Run(app *App) error {
	ctx := context.Background()
	base, err := resolveDirectionsCommon(ctx, app, c.DirectionsCommon)
	if err != nil {
		return err
	}
	resp, err := app.client.DirectionsBicycling(ctx, amapclient.BicyclingRequest{
		DirectionsRequest: base,
		AlternativeRoute:  c.AlternativeRoute,
	})
	if err != nil {
		return err
	}
	return app.writeDirections(resp)
}

func (c *DirectionsElectrobikeCmd) Run(app *App) error {
	ctx := context.Background()
	base, err := resolveDirectionsCommon(ctx, app, c.DirectionsCommon)
	if err != nil {
		return err
	}
	resp, err := app.client.DirectionsElectrobike(ctx, amapclient.ElectrobikeRequest{
		DirectionsRequest: base,
		AlternativeRoute:  c.AlternativeRoute,
	})
	if err != nil {
		return err
	}
	return app.writeDirections(resp)
}

func (c *DirectionsTransitCmd) Run(app *App) error {
	ctx := context.Background()
	base, err := resolveDirectionsCommon(ctx, app, c.DirectionsCommon)
	if err != nil {
		return err
	}
	resp, err := app.client.DirectionsTransit(ctx, amapclient.TransitRequest{
		DirectionsRequest: base,
		City1:             c.City1,
		City2:             c.City2,
		Strategy:          c.Strategy,
		AlternativeRoute:  c.AlternativeRoute,
	})
	if err != nil {
		return err
	}
	return app.writeDirections(resp)
}

func resolveDirectionsCommon(ctx context.Context, app *App, c DirectionsCommon) (amapclient.DirectionsRequest, error) {
	origin, err := resolveLocation(ctx, app, c.Origin)
	if err != nil {
		return amapclient.DirectionsRequest{}, err
	}
	destination, err := resolveLocation(ctx, app, c.Destination)
	if err != nil {
		return amapclient.DirectionsRequest{}, err
	}
	return amapclient.DirectionsRequest{
		Origin:      origin,
		Destination: destination,
	}, nil
}

func (a *App) writeDirections(response amapclient.DirectionsResponse) error {
	if a.json {
		encoded, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		_, _ = fmt.Fprintln(a.out, string(encoded))
		return nil
	}
	_, _ = fmt.Fprint(a.out, renderDirections(a.color, response))
	return nil
}

func (c *WeatherCmd) Run(app *App) error {
	req := amapclient.WeatherRequest{
		City:       c.City,
		Extensions: c.Extensions,
	}
	response, err := app.client.Weather(context.Background(), req)
	if err != nil {
		return err
	}
	return app.writeWeather(response)
}

func (a *App) writeWeather(response amapclient.WeatherResponse) error {
	if a.json {
		encoded, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		_, _ = fmt.Fprintln(a.out, string(encoded))
		return nil
	}
	_, _ = fmt.Fprint(a.out, renderWeather(a.color, response))
	return nil
}

func (a *App) writeTips(response amapclient.InputTipsResponse) error {
	if a.json {
		encoded, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		_, _ = fmt.Fprintln(a.out, string(encoded))
		return nil
	}
	_, _ = fmt.Fprint(a.out, renderTips(a.color, response))
	return nil
}

func (a *App) writePOIs(response amapclient.NearbySearchResponse) error {
	if a.json {
		encoded, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		_, _ = fmt.Fprintln(a.out, string(encoded))
		return nil
	}
	_, _ = fmt.Fprint(a.out, renderNearby(a.color, response))
	return nil
}
