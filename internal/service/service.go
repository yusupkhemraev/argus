package service

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	linuxServiceName = "argus"
	macPlistLabel    = "io.github.yusupkhemraev.argus"
	macPlistPath     = "~/Library/LaunchAgents/io.github.yusupkhemraev.argus.plist"
)

func Start() error   { return control("start") }
func Stop() error    { return control("stop") }
func Restart() error { return control("restart") }

func Status() (string, error) {
	switch runtime.GOOS {
	case "linux":
		out, err := run("systemctl", "status", linuxServiceName, "--no-pager", "-l")
		return out, err
	case "darwin":
		out, _ := run("launchctl", "list", macPlistLabel)
		if out == "" || strings.Contains(out, "Could not find") {
			return "stopped", nil
		}
		return out, nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func control(action string) error {
	switch runtime.GOOS {
	case "linux":
		return runSudo("systemctl", action, linuxServiceName)
	case "darwin":
		switch action {
		case "start":
			return runCmd("launchctl", "load", "-w", expandHome(macPlistPath))
		case "stop":
			return runCmd("launchctl", "unload", expandHome(macPlistPath))
		case "restart":
			_ = runCmd("launchctl", "unload", expandHome(macPlistPath))
			return runCmd("launchctl", "load", "-w", expandHome(macPlistPath))
		}
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

func runSudo(name string, args ...string) error {
	return runCmd("sudo", append([]string{name}, args...)...)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := exec.Command("sh", "-c", "echo $HOME").Output()
		return strings.TrimSpace(string(home)) + path[1:]
	}
	return path
}