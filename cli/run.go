package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	amapclient "github.com/minggeorgelei/AMAP-client"
)

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
	req := amapclient.NearbySearchRequest{
		Location: c.Location,
		Keywords: c.Keywords,
		Radius:   c.Radius,
		SortRule: c.SortRule,
		Limit:    c.Limit,
		Filter: amapclient.NearbySearchFilter{
			MinCost:   c.MinCost,
			MaxCost:   c.MaxCost,
			MinRating: c.MinRating,
		},
	}
	response, err := app.client.NearbySearch(context.Background(), req)
	if err != nil {
		return err
	}
	return app.writePOIs(response)
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
