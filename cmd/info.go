package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/urfave/cli"
	"github.com/warpdl/warpdl/pkg/warpcli"
	"github.com/warpdl/warpdl/pkg/warplib"
)

func info(ctx *cli.Context) error {
	url := ctx.Args().First()
	if url == "" {
		return printErrWithCmdHelp(
			ctx,
			errors.New("no url provided"),
		)
	} else if url == "help" {
		return cli.ShowCommandHelp(ctx, ctx.Command.Name)
	}
	fmt.Printf("%s: fetching details, please wait...\n", ctx.App.HelpName)
	var headers warplib.Headers
	if userAgent != "" {
		headers = warplib.Headers{{
			Key: warplib.USER_AGENT_KEY, Value: getUserAgent(userAgent),
		}}
	}
	d, err := warplib.NewDownloader(
		&http.Client{},
		url,
		&warplib.DownloaderOpts{
			Headers:   headers,
			SkipSetup: true,
		},
	)
	if err != nil {
		printRuntimeErr(ctx, "info", "new_downloader", err)
		return nil
	}
	fName := d.GetFileName()
	if fName == "" {
		fName = "not-defined"
	}
	fmt.Printf(`
File Info
Name`+"\t"+`: %s
Size`+"\t"+`: %s
`, fName, d.GetContentLengthAsString())
	return nil
}

func list(ctx *cli.Context) error {
	if ctx.Args().First() == "help" {
		return cli.ShowCommandHelp(ctx, ctx.Command.Name)
	}
	client, err := warpcli.NewClient()
	if err != nil {
		printRuntimeErr(ctx, "list", "new_client", err)
		return nil
	}
	l, err := client.List(&warpcli.ListOpts{
		ShowCompleted: showCompleted || showAll,
		ShowPending:   showPending || showAll,
	})
	if err != nil {
		printRuntimeErr(ctx, "list", "get_list", err)
		return nil
	}
	fback := func() error {
		fmt.Println("warp: no downloads found")
		return nil
	}
	if len(l.Items) == 0 {
		return fback()
	}
	txt := "Here are your downloads:"
	txt += "\n\n------------------------------------------------------"
	txt += "\n|Num|\t         Name         | Unique Hash | Status |"
	txt += "\n|---|-------------------------|-------------|--------|"
	var i int
	for _, item := range l.Items {
		if !showHidden && (item.Hidden || item.Children) {
			continue
		}
		i++
		name := item.Name
		n := len(name)
		switch {
		case n > 23:
			name = name[:20] + "..."
		case n < 23:
			name = beaut(name, 23)
		}
		perc := fmt.Sprintf(`%d%%`, item.GetPercentage())
		txt += fmt.Sprintf("\n| %d | %s |   %s  |  %s  |", i, name, item.Hash, beaut(perc, 4))
	}
	if i == 0 {
		return fback()
	}
	txt += "\n------------------------------------------------------"
	fmt.Println(txt)
	return nil
}

func beaut(s string, n int) (b string) {
	n1 := len(s)
	x := n - n1
	x1 := x / 2
	w := string(
		replic(' ', x1),
	)
	b = w
	b += s
	b += w
	if x%2 != 0 {
		b += " "
	}
	return
}

func replic[aT any](v aT, n int) []aT {
	a := make([]aT, n)
	for i := range a {
		a[i] = v
	}
	return a
}