/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"gopkg.in/yaml.v3"
)

type ConfigVersion int

const (
	FirstConfigVersion ConfigVersion = 1

	CurrentConfigVersion ConfigVersion = FirstConfigVersion
)

const ConfigPathTemplate = "%s/wailsterm/config.yaml"

type WindowTheme string

const (
	WindowThemeAuto  WindowTheme = "auto"
	WindowThemeLight WindowTheme = "light"
	WindowThemeDark  WindowTheme = "dark"
)

func (t WindowTheme) AsMacAppearanceType() mac.AppearanceType {
	switch t {
	case WindowThemeAuto:
		return mac.DefaultAppearance
	case WindowThemeLight:
		return mac.NSAppearanceNameAqua
	case WindowThemeDark:
		return mac.NSAppearanceNameDarkAqua
	default:
		panic("unknown WindowTheme")
	}
}

type TerminalTheme string

const (
	TerminalThemeOneHalf TerminalTheme = "OneHalf"
)

var TerminalThemesEnum = []struct {
	Value  TerminalTheme
	TSName string
}{
	{TerminalThemeOneHalf, string(TerminalThemeOneHalf)},
}

type TerminalFontFamily string

const (
	TerminalFontFamilyFiraCode TerminalFontFamily = "FiraCode"
)

var TerminalFontsEnum = []struct {
	Value  TerminalFontFamily
	TSName string
}{
	{TerminalFontFamilyFiraCode, string(TerminalFontFamilyFiraCode)},
}

type ShellEnvConfig struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ShellEnvsConfig []ShellEnvConfig

func (envs ShellEnvsConfig) AsStrings() []string {
	var result []string
	for _, env := range envs {
		result = append(result, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	return result
}

type ShellConfig struct {
	Path string          `yaml:"path"`
	Args []string        `yaml:"args"`
	Envs ShellEnvsConfig `yaml:"envs"`
}

type WindowSizeConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type WindowConfig struct {
	Theme WindowTheme      `yaml:"theme"`
	Size  WindowSizeConfig `yaml:"size"`
}

type TerminalFontConfig struct {
	Family     TerminalFontFamily `yaml:"family"`
	Size       int                `yaml:"size"`
	Weight     int                `yaml:"weight"`
	WeightBold int                `yaml:"weight_bold"`
}

type TerminalConfig struct {
	Theme TerminalTheme      `yaml:"theme"`
	Font  TerminalFontConfig `yaml:"font"`
}

type ApplicationConfig struct {
	SingleInstance *bool `yaml:"single_instance"`
}

func (a ApplicationConfig) IsSingleInstance() bool {
	return a.SingleInstance != nil && *a.SingleInstance
}

func (a ApplicationConfig) AsSingleInstanceLock(app *App) *options.SingleInstanceLock {
	if a.IsSingleInstance() {
		return &options.SingleInstanceLock{
			UniqueId:               "d4366eb9-0a70-48ed-a97d-8b4035efa7fc",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		}
	}

	return nil
}

type Config struct {
	Version     ConfigVersion     `yaml:"version"`
	Application ApplicationConfig `yaml:"application"`
	Shell       ShellConfig       `yaml:"shell"`
	Window      WindowConfig      `yaml:"window"`
	Terminal    TerminalConfig    `yaml:"terminal"`
}

func defaultConfig() Config {
	singleInstance := true
	shellConfig := ShellConfig{
		Path: "/usr/bin/env",
		Args: []string{"/bin/zsh", "--login"},
		Envs: []ShellEnvConfig{
			{Name: "TERM", Value: "xterm-256color"},
			{Name: "LC_CTYPE", Value: "UTF-8"},
		},
	}
	if currentPlatform() == PlatformWindows {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = "c:/windows"
		}
		shellConfig.Path = filepath.Join(systemRoot, "System32/cmd.exe")
		shellConfig.Args = []string{"/k"}
	} else if currentPlatform() == PlatformLinux {
		if _, err := os.Stat("/bin/bash"); os.IsNotExist(err) {
			shellConfig.Path = "/bin/sh"
		} else {
			shellConfig.Path = "/bin/bash"
		}
		shellConfig.Args = []string{}
	}
	return Config{
		Version: CurrentConfigVersion,
		Application: ApplicationConfig{
			SingleInstance: &singleInstance,
		},
		Shell: shellConfig,
		Window: WindowConfig{
			Theme: WindowThemeAuto,
			Size: WindowSizeConfig{
				Width:  600,
				Height: 410,
			},
		},
		Terminal: TerminalConfig{
			Theme: TerminalThemeOneHalf,
			Font: TerminalFontConfig{
				Family:     TerminalFontFamilyFiraCode,
				Size:       15,
				Weight:     400,
				WeightBold: 600,
			},
		},
	}
}

func getConfigPath() (string, error) {
	var basePath string

	switch currentPlatform() {
	case PlatformMacOs:
		var err error
		basePath, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
		basePath = strings.TrimRight(basePath, string(os.PathSeparator))
		basePath = basePath + string(os.PathSeparator) + ".config"

	default:
		var err error
		basePath, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}

	basePath = strings.TrimRight(basePath, string(os.PathSeparator))

	return fmt.Sprintf(ConfigPathTemplate, basePath), nil
}

func writeDefaultConfig() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0740)
	if err != nil {
		return err
	}

	c := defaultConfig()
	yamlFile, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	_, err = io.WriteString(f, string(yamlFile))
	return err
}

func loadConfig() (Config, error) {
	p, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(p)
	if os.IsNotExist(err) {
		err = writeDefaultConfig()
		if err != nil {
			return Config{}, err
		}

		return defaultConfig(), nil
	} else if err != nil {
		return Config{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	err = mergo.Merge(&c, defaultConfig(), mergo.WithoutDereference)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
