package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New", keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:file:new")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Open...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:file:open")
	})
	fileMenu.AddText("Save", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:file:save")
	})
	fileMenu.AddText("Save As...", keys.CmdOrCtrl("shift+s"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:file:save-as")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	editMenu := appMenu.AddSubmenu("Edit")
	editMenu.AddText("Undo", keys.CmdOrCtrl("z"), nil) // Native handler
	editMenu.AddText("Redo", keys.CmdOrCtrl("y"), nil) // Native handler
	editMenu.AddSeparator()
	editMenu.AddText("Cut", keys.CmdOrCtrl("x"), nil)  // Native handler
	editMenu.AddText("Copy", keys.CmdOrCtrl("c"), nil) // Native handler
	editMenu.AddText("Paste", keys.CmdOrCtrl("v"), nil) // Native handler

	layoutMenu := appMenu.AddSubmenu("Layout")

	themesMenu := layoutMenu.AddSubmenu("Themes")
	themeIDs := []string{"midnight", "parchment", "blossom", "studio", "crimson", "seafoam", "ember", "viola", "moss", "frost"}
	for _, id := range themeIDs {
		themeID := id
		themesMenu.AddText(id, nil, func(_ *menu.CallbackData) {
			runtime.EventsEmit(app.ctx, "menu:layout:theme:"+themeID)
		})
	}

	fontsMenu := layoutMenu.AddSubmenu("Fonts")
	fontIDs := []string{"literary", "editorial", "typewriter", "playfair", "inter", "merriweather", "monoton", "bebas"}
	for _, id := range fontIDs {
		fontID := id
		fontsMenu.AddText(id, nil, func(_ *menu.CallbackData) {
			runtime.EventsEmit(app.ctx, "menu:layout:font:"+fontID)
		})
	}

	layoutMenu.AddSeparator()

	sizeMenu := layoutMenu.AddSubmenu("Font Size")
	sizes := []struct{ label, value string }{
		{"Small", "small"},
		{"Normal", "normal"},
		{"Large", "large"},
		{"Huge", "huge"},
	}
	for _, s := range sizes {
		sizeValue := s.value
		sizeMenu.AddText(s.label, nil, func(_ *menu.CallbackData) {
			runtime.EventsEmit(app.ctx, "menu:layout:font-size:"+sizeValue)
		})
	}

	spacingMenu := layoutMenu.AddSubmenu("Spacing")
	spacings := []struct{ label, value string }{
		{"Tight", "tight"},
		{"Comfortable", "comfortable"},
		{"Relaxed", "relaxed"},
	}
	for _, s := range spacings {
		spacingValue := s.value
		spacingMenu.AddText(s.label, nil, func(_ *menu.CallbackData) {
			runtime.EventsEmit(app.ctx, "menu:layout:spacing:"+spacingValue)
		})
	}

	settingsMenu := appMenu.AddSubmenu("Settings")
	settingsMenu.AddText("Configure API Key...", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:settings:configure-api-key")
	})

	err := wails.Run(&options.App{
		Title:     "PandaWriter",
		Width:     1440,
		Height:    900,
		MinWidth:  960,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:   &options.RGBA{R: 20, G: 20, B: 30, A: 1},
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.ERROR,
		Menu:               appMenu,
		Debug: options.Debug{
			OpenInspectorOnStartup: true,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
			&app.credentials,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}