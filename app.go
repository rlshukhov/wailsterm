/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
)

type App struct {
	ctx  context.Context
	term *term
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	t, err := initTerm(func() {
		a.Quit()
	})
	if err != nil {
		log.Println("Init term error:", err)
		a.Quit()
	}
	a.term = t
}

func (a *App) GetWebsocketUrl() string {
	return a.term.GetWsUrl()
}

func (a *App) GetPlatform() Platform {
	return currentPlatform()
}

func (a *App) GetTerminalTheme() TerminalTheme {
	return config.Terminal.Theme
}

func (a *App) GetTerminalFontConfig() TerminalFontConfig {
	return config.Terminal.Font
}

func (a *App) SetPtySize(rows, cols int) {
	if err := a.term.Resize(rows, cols); err != nil {
		log.Println("Pty resize error:", err)
	}
}

func (a *App) Quit() {
	log.Println("manually quit app")

	err := a.term.Close()
	if err != nil {
		log.Println("close term error:", err)
	}

	runtime.Quit(a.ctx)
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	log.Println("second instance launch")

	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
}
