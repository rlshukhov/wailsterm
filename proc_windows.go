//go:build windows

package main

import "syscall"

func createSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
