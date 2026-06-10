package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func openBrowser(url, browser string) {
	var cmd *exec.Cmd

	if browser != "" {
		cmd = exec.Command(browser, url)
	} else {
		switch runtime.GOOS {
		case "linux":
			if isWSL() {
				cmd = exec.Command(findCmdExe(), "/c", "start", url)
			} else {
				cmd = exec.Command("xdg-open", url)
			}
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd.exe", "/c", "start", url)
		default:
			return
		}
	}

	if err := cmd.Start(); err != nil {
		log.Printf("failed to open browser: %v", err)
	}
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte("microsoft")) ||
		bytes.Contains(data, []byte("WSL"))
}

func findCmdExe() string {
	// WSL defaut use windowsPATH
	if path, err := exec.LookPath("cmd.exe"); err == nil {
		return path
	}

	// Fallback
	return "/mnt/c/Windows/System32/cmd.exe"
}
