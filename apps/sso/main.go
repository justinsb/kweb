package main

import (
	"embed"

	"github.com/justinsb/kweb"
)

//go:embed pages
var pages embed.FS

func main() {
	opt := kweb.NewOptions("kweb-sso-system")
	opt.Server.Pages.Base = pages
	app := kweb.NewApp(opt)
	app.RunFromMain()
}
