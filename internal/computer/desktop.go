package computer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type DesktopManager struct {
	activeProcesses []*exec.Cmd
}

func NewDesktopManager() *DesktopManager {
	return &DesktopManager{
		activeProcesses: make([]*exec.Cmd, 0),
	}
}

func (dm *DesktopManager) Launch(appName string, args []string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", append([]string{"/c", "start", appName}, args...)...)
	case "darwin":
		cmd = exec.Command("open", append([]string{"-a", appName}, args...)...)
	default:
		cmd = exec.Command(appName, args...)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch %s: %w", appName, err)
	}
	dm.activeProcesses = append(dm.activeProcesses, cmd)
	return nil
}

func (dm *DesktopManager) RunCommand(command string) (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", command)
	default:
		cmd = exec.Command("sh", "-c", command)
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("run command: %w\nstderr: %s", err, stderr.String())
	}
	return out.String(), nil
}

func (dm *DesktopManager) FileOp(op, path, content string) error {
	switch op {
	case "read":
		_, err := os.ReadFile(path)
		return err
	case "write":
		return os.WriteFile(path, []byte(content), 0644)
	case "delete":
		return os.Remove(path)
	case "mkdir":
		return os.MkdirAll(path, 0755)
	case "list":
		_, err := os.ReadDir(path)
		return err
	default:
		return fmt.Errorf("unknown file operation: %s", op)
	}
}

func (dm *DesktopManager) OpenTerminal(dir string) error {
	app := ""
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		app = "cmd.exe"
		args = []string{"/k", fmt.Sprintf("cd /d %s", dir)}
	case "darwin":
		app = "Terminal"
		args = []string{dir}
	default:
		app = "x-terminal-emulator"
		args = []string{fmt.Sprintf("--working-directory=%s", dir)}
	}
	return dm.Launch(app, args)
}

func (dm *DesktopManager) OpenFileManager(dir string) error {
	app := ""
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		app = "explorer"
		args = []string{dir}
	case "darwin":
		app = "Finder"
		args = []string{dir}
	default:
		app = "nautilus"
		args = []string{dir}
	}
	return dm.Launch(app, args)
}

func (dm *DesktopManager) OpenBrowser(url string) error {
	app := ""
	args := []string{}
	switch runtime.GOOS {
	case "windows":
		app = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		app = "open"
		args = []string{url}
	default:
		app = "xdg-open"
		args = []string{url}
	}
	return dm.Launch(app, args)
}

func (dm *DesktopManager) FindExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	return path, nil
}

func (dm *DesktopManager) KillAll() {
	for _, cmd := range dm.activeProcesses {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
	dm.activeProcesses = nil
}

func (dm *DesktopManager) PlaywrightLaunch(browserType, scriptPath string, args []string) (string, error) {
	playwrightCmd := "npx"
	playwrightArgs := []string{"playwright"}

	switch browserType {
	case "chromium":
		playwrightArgs = append(playwrightArgs, "run", scriptPath)
	case "firefox":
		playwrightArgs = append(playwrightArgs, "run", "--browser=firefox", scriptPath)
	case "webkit":
		playwrightArgs = append(playwrightArgs, "run", "--browser=webkit", scriptPath)
	default:
		return "", fmt.Errorf("unknown playwright browser: %s", browserType)
	}

	playwrightArgs = append(playwrightArgs, args...)
	return dm.RunCommand(strings.Join(append([]string{playwrightCmd}, playwrightArgs...), " "))
}

func (dm *DesktopManager) PuppeteerLaunch(scriptPath string, args []string) (string, error) {
	puppeteerArgs := append([]string{"puppeteer"}, args...)
	return dm.RunCommand(strings.Join(append([]string{"npx", scriptPath}, puppeteerArgs...), " "))
}
