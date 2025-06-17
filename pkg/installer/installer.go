package installer

import (
	"os/exec"
	"runtime"
)

func Install(app AppConfig) error {
	var cmd *exec.Cmd

	if app.URL != "" {
		cmd = exec.Command(app.Cmd, app.Args...)
	} else if app.Path != "" {
		if runtime.GOOS == "windows" {
			cmd = exec.Command(app.Path, app.Args...)
		} else {
			cmd = exec.Command("sh", "-c", app.Path)
		}
	}
	return cmd.Run()
}
