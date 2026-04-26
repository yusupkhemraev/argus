package service

import (
	"fmt"
	"os/exec"
	"strings"
)

const serviceName = "argus"

func Start() error   { return control("start") }
func Stop() error    { return control("stop") }
func Restart() error { return control("restart") }

func Status() (string, error) {
	return run("systemctl", "status", serviceName, "--no-pager", "-l")
}

func control(action string) error {
	return runSudo("systemctl", action, serviceName)
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runCmd(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

func runSudo(name string, args ...string) error {
	return runCmd("sudo", append([]string{name}, args...)...)
}
