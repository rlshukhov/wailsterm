/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/creack/pty"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var ptmx *os.File
var app *App

func main() {
	// Create arbitrary command.
	c := exec.Command("/usr/bin/env", "/bin/zsh", "--login")
	home := os.Getenv("HOME")
	c.Env = append(os.Environ(), "TERM=xterm-256color", "LC_CTYPE=UTF-8")
	c.Dir = home

	// Start the command with a pty.
	var err error
	ptmx, err = pty.Start(c)
	if err != nil {
		log.Println("Start pty error:", err)
		return
	}
	defer func() {
		log.Println("Close pty")

		err := ptmx.Close()
		if err != nil {
			log.Println("Close PTY error:", err)
		}

		ptmx = nil
	}()

	// Create an instance of the app structure
	app = NewApp(ptmx)

	if err := pty.Setsize(ptmx, app.ptySize); err != nil {
		log.Println("error resizing pty:", err)
		return
	}

	go func() {
		http.HandleFunc(app.wsPath+app.accessToken, sshHandler)
		log.Fatal(http.ListenAndServe(app.wsListenAddress, nil))
	}()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "WailsTerm",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			Appearance:           mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "WailsTerm",
				Message: "© 2024 Lane Shukhov",
				Icon:    icon,
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			DisableWindowIcon:    true,
			DisablePinchZoom:     true,
			BackdropType:         windows.Acrylic,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "d4366eb9-0a70-48ed-a97d-8b4035efa7fc",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func sshHandler(w http.ResponseWriter, r *http.Request) {
	if ptmx == nil {
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer func() {
		err := wsConn.Close()
		if err != nil {
			log.Println("Close websocket connection error:", err)
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			if ptmx == nil {
				return
			}

			n, err := io.LimitReader(ptmx, 1024).Read(buf)

			if err != nil {
				if err != io.EOF {
					log.Println("Read from PTY stdout error:", err)
				}

				app.Quit()

				return
			}
			if n > 0 {
				err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
				if err != nil {
					log.Println("Write to WebSocket error:", err)
					return
				}
			}
		}
	}()

	for {
		messageType, p, err := wsConn.ReadMessage()

		if err != nil {
			if err != io.EOF {
				log.Println("Read from WebSocket error:", err)
			}
			return
		}
		if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
			if ptmx == nil {
				return
			}

			_, err = io.Copy(ptmx, bytes.NewReader(p))
			if err != nil {
				log.Println("Write to PTY stdin error:", err)
				return
			}
		}
	}
}
