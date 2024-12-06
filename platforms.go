/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import "runtime"

type Platform string

const (
	PlatformMacOs   Platform = "darwin"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

var PlatformsEnum = []struct {
	Value  Platform
	TSName string
}{
	{PlatformMacOs, "MacOs"},
	{PlatformLinux, "Linux"},
	{PlatformWindows, "Windows"},
}

func currentPlatform() Platform {
	switch runtime.GOOS {
	case string(PlatformMacOs):
		return PlatformMacOs
	case string(PlatformLinux):
		return PlatformLinux
	case string(PlatformWindows):
		return PlatformWindows
	default:
		panic("unsupported platform")
	}
}
