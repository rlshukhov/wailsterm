/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/src/assets/images/logo-256.png
var icon []byte

var config Config

func main() {
	var err error
	config, err = loadConfig()
	if err != nil {
		panic(err)
	}

	runApp()
}

func runApp() {
	app := NewApp()

	appMenu := menu.NewMenu()
	if currentPlatform() == PlatformMacOs {
		appMenu.Append(menu.AppMenu())
	}

	FileMenu := appMenu.AddSubmenu("File")
	if !config.Application.IsSingleInstance() {
		FileMenu.AddText("New window", keys.CmdOrCtrl("n"), spawnNewApp)
	}
	FileMenu.AddText("Clear caches", nil, app.clearCaches)

	if currentPlatform() == PlatformMacOs {
		appMenu.Append(menu.EditMenu())
		appMenu.Append(menu.WindowMenu())
	}

	err := wails.Run(&options.App{
		Title:  "WailsTerm",
		Width:  config.Window.Size.Width,
		Height: config.Window.Size.Height,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Menu:      appMenu,
		Bind: []any{
			app,
		},
		EnumBind: []any{
			PlatformsEnum,
			TerminalThemesEnum,
			TerminalFontsEnum,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			Appearance:           config.Window.Theme.AsMacAppearanceType(),
			About: &mac.AboutInfo{
				Title: "WailsTerm",
				Message: strings.Join([]string{
					"© 2024 Lane Shukhov",
				}, "\n"),
				Icon: icon,
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    true,
			DisablePinchZoom:     true,
			BackdropType:         windows.Acrylic,
		},
		SingleInstanceLock: config.Application.AsSingleInstanceLock(app),
	})

	if err != nil {
		log.Println("Error:", err.Error())
	}
}

func spawnNewApp(data *menu.CallbackData) {
	exePath, err := os.Executable()
	if err != nil {
		log.Println("Locate current executable error:", err)
		return
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		log.Println("Can't get current executable absolute path:", err)
		return
	}

	cmd := exec.Command(exePath)
	cmd.Stdin = nil

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err = cmd.Start()
	if err != nil {
		log.Println("Can't spawn new executable copy:", err)
		return
	}
}
