//go:build windows
// +build windows

package main

import (
	"fmt"
	"os/exec"
)

func shutdownNow(reboot bool) error {
	var command = "/s"
	if reboot {
		command = "/r"
	}

	fmt.Println("Shutdown win. Reboot:", command)
	if err := exec.Command("cmd", "/C", "shutdown", "/t", "0", command).Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
		return err
	}

	return nil
}
