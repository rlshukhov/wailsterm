/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"os"
)

// App struct
type App struct {
	ctx             context.Context
	wsListenAddress string
	wsPath          string
	accessToken     string
	ptmx            *os.File
	ptySize         *pty.Winsize
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(s int) string {
	b, err := GenerateRandomBytes(s)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

func NewApp(ptmx *os.File) *App {
	return &App{
		wsListenAddress: "127.0.0.1:62103",
		wsPath:          "/ws/pty/",
		accessToken:     generateRandomString(32),
		ptySize: &pty.Winsize{
			Rows: 10,
			Cols: 60,
		},
		ptmx: ptmx,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetWebsocketUrl() string {
	return "ws://" + a.wsListenAddress + a.wsPath + a.accessToken
}

func (a *App) SetPtySize(rows, cols int) {
	newRows := uint16(rows)
	newCols := uint16(cols)

	if a.ptySize.Rows == newRows && a.ptySize.Cols == newCols {
		return
	}

	a.ptySize.Rows = newRows
	a.ptySize.Cols = newCols

	err := pty.Setsize(a.ptmx, a.ptySize)
	if err != nil {
		log.Println("error resizing pty:", err)
	}
}

func (a *App) Quit() {
	log.Println("manually quit app")
	runtime.Quit(a.ctx)
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	log.Println("second instance launch")

	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
}
