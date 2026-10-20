package app

import (
	"io/fs"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// Run starts the Wails window with the embedded frontend assets and binds the
// App object (exposed to JS as window.go.app.App.*). Window options live here
// so main.go only has to embed the frontend bundle (web/dist) and call Run.
func Run(assets fs.FS) error {
	app := NewApp()
	return wails.Run(&options.App{
		Title:         "hatch",
		Width:         1024,
		Height:        768,
		MinWidth:      900,
		MinHeight:     640,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			Preferences: &mac.Preferences{
				FullscreenEnabled: mac.Enabled,
			},
			DisableZoom: false,
		},
	})
}
