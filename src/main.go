// Hatch backend entry point (Go module root). The frontend lives in
// ../frontend; `npm run build` there writes its bundle into web/dist so it can
// be embedded here (//go:embed cannot reference parent directories).
package main

import (
	"embed"

	"hatch/app"
)

//go:embed all:web/dist
var assets embed.FS

func main() {
	if err := app.Run(assets); err != nil {
		println("Error:", err.Error())
	}
}
