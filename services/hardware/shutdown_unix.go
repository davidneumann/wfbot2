//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os/exec"
)

func shutdownNow(reboot bool) error {
	var command = "-h"
	if reboot {
		command = "-r"
	}
	if err := exec.Command("shutdown", command, "now").Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
		return err
	}

	return nil
}
