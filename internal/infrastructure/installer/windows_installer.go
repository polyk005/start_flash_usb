package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/polyk005/start_flash/internal/domain"
	"github.com/pkg/errors"
)

type WindowsInstaller struct {
	logger          domain.Logger
	installBasePath string
}

func NewWindowsInstaller(logger domain.Logger, installBasePath string) *WindowsInstaller {
	return &WindowsInstaller{
		logger:          logger,
		installBasePath: installBasePath,
	}
}

func (i *WindowsInstaller) Install(ctx context.Context, app *domain.App) (*domain.Installation, error) {
	logger := i.logger.WithFields(map[string]interface{}{
		"app_id":   app.ID,
		"app_name": app.Name,
	})

	logger.Info("Starting Windows installation")

	if err := i.validateApp(app); err != nil {
		return nil, errors.Wrap(err, "invalid app configuration")
	}

	installerPath, err := i.getInstallerPath(app)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get installer path")
	}

	logger.WithField("installer_path", installerPath).Debug("Using installer path")

	installation := &domain.Installation{
		AppID:     app.ID,
		Status:    domain.StatusInstalling,
		StartTime: time.Now(),
	}

	if err := i.executeInstallation(ctx, app, installerPath, installation); err != nil {
		installation.Status = domain.StatusFailed
		installation.Error = err.Error()
		installation.EndTime = &[]time.Time{time.Now()}[0]
		installation.Duration = installation.EndTime.Sub(installation.StartTime)
		return installation, err
	}

	installation.Status = domain.StatusSuccess
	installation.EndTime = &[]time.Time{time.Now()}[0]
	installation.Duration = installation.EndTime.Sub(installation.StartTime)

	logger.Info("Windows installation completed successfully")
	return installation, nil
}

func (i *WindowsInstaller) validateApp(app *domain.App) error {
	if app.Name == "" {
		return errors.New("app name is required")
	}

	if app.Path == "" && app.URL == "" {
		return errors.New("either path or URL must be specified")
	}

	if len(app.Args) == 0 {
		i.logger.WithField("app_name", app.Name).Warn("No installation arguments specified")
	}

	return nil
}

func (i *WindowsInstaller) getInstallerPath(app *domain.App) (string, error) {
	if app.Path != "" {
		if filepath.IsAbs(app.Path) {
			return app.Path, nil
		}
		absPath, err := filepath.Abs(app.Path)
		if err != nil {
			return "", errors.Wrapf(err, "failed to resolve absolute path for %s", app.Path)
		}
		return absPath, nil
	}

	if app.URL != "" {
		fileName := i.getFileNameFromURL(app.URL)
		// Используем абсолютный путь к папке apps
		appsDir, err := filepath.Abs("./apps")
		if err != nil {
			return "", errors.Wrapf(err, "failed to resolve apps directory path")
		}
		installerPath := filepath.Join(appsDir, fileName)

		absPath, err := filepath.Abs(installerPath)
		if err != nil {
			return "", errors.Wrapf(err, "failed to resolve absolute path for %s", installerPath)
		}

		// Проверяем существование файла
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			// Также проверяем версию с .exe расширением
			exePath := absPath
			if !strings.HasSuffix(strings.ToLower(absPath), ".exe") {
				exePath = absPath + ".exe"
			}
			
			if _, err := os.Stat(exePath); os.IsNotExist(err) {
				// Файл не найден, скачиваем
				i.logger.WithField("url", app.URL).Info("Installer not found locally, downloading...")
				if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
					return "", errors.Wrapf(err, "failed to create directory %s", filepath.Dir(absPath))
				}
				// Download and maybe rename
				absPath, err = i.downloadFile(app.URL, absPath)
				if err != nil {
					return "", errors.Wrapf(err, "failed to download installer from %s", app.URL)
				}
			} else {
				// Файл с .exe найден
				absPath = exePath
			}
		}
		return absPath, nil
	}

	return "", errors.New("no valid installer path found")
}

func (i *WindowsInstaller) getFileNameFromURL(url string) string {
	// Убираем параметры запроса (все после ?)
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}
	
	// Получаем базовое имя файла
	fileName := filepath.Base(url)
	
	// Если имя файла пустое или не имеет расширения, используем имя приложения
	if fileName == "" || fileName == "." || fileName == "/" || fileName == "download" {
		// Генерируем имя файла на основе домена и пути
		if strings.Contains(url, "discord.com") {
			return "discord_setup.exe"
		} else if strings.Contains(url, "visualstudio.com") {
			return "vscode_setup.exe"
		} else if strings.Contains(url, "sourceforge.net") {
			return "qbittorrent_setup.exe"
		} else if strings.Contains(url, "telegram.org") {
			return "telegram_setup.exe"
		} else if strings.Contains(url, "google.com") {
			return "chrome_setup.exe"
		} else if strings.Contains(url, "rarlab.com") {
			return "winrar-x64-711ru.exe"
		} else if strings.Contains(url, "steamstatic.com") {
			return "steam_setup.exe"
		} else if strings.Contains(url, "epicgames.com") {
			return "epic_launcher_setup.exe"
		} else if strings.Contains(url, "docker.com") {
			return "docker_desktop_setup.exe"
		} else {
			return "installer.exe"
		}
	}
	
	return fileName
}

func (i *WindowsInstaller) downloadFile(url, filepath string) (string, error) {
	logger := i.logger.WithFields(map[string]interface{}{
		"url":      url,
		"filepath": filepath,
	})
	
	logger.Info("Downloading installer file")
	
	// Создаем HTTP-клиент с таймаутом
	client := &http.Client{
		Timeout: 5 * time.Minute, // 5 минут таймаут
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "failed to download file from %s", url)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}
	
	file, err := os.Create(filepath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create file %s", filepath)
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "failed to write file %s", filepath)
	}
	
	logger.Info("File downloaded successfully")

	// Если файл не заканчивается на .exe, переименовать
	if !strings.HasSuffix(strings.ToLower(filepath), ".exe") {
		newPath := filepath + ".exe"
		err := os.Rename(filepath, newPath)
		if err != nil {
			return "", errors.Wrapf(err, "failed to rename file to %s", newPath)
		}
		logger.WithField("new_path", newPath).Info("Renamed downloaded file to .exe")
		return newPath, nil
	}
	return filepath, nil
}

func (i *WindowsInstaller) executeInstallation(ctx context.Context, app *domain.App, installerPath string, installation *domain.Installation) error {
	logger := i.logger.WithFields(map[string]interface{}{
		"app_id":         app.ID,
		"installer_path": installerPath,
		"args":           app.Args,
	})

	logger.Info("Executing installation command")

	args := app.Args
	if i.installBasePath != "" && app.InstallDirArgTpl != "" {
		targetDir := filepath.Join(i.installBasePath, app.Name)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			logger.WithField("error", err).Error("Failed to create target installation directory")
		} else {
			dirArg := fmt.Sprintf(app.InstallDirArgTpl, targetDir)
			args = append(args, dirArg)
			logger.WithField("dir_arg", dirArg).Info("Added custom installation directory argument")
		}
	}

	cmd := exec.CommandContext(ctx, "cmd", "/C", "start", "/wait", installerPath)
	cmd.Args = append(cmd.Args, args...)

	cmd.Dir = filepath.Dir(installerPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithField("error", err).WithField("output", string(output)).Error("Installation command failed")
		return errors.Wrapf(err, "installation failed: %s", string(output))
	}

	logger.WithField("output", string(output)).Debug("Installation command completed")
	
	if len(output) > 0 {
		installation.Logs = append(installation.Logs, string(output))
	}

	return nil
}

type WindowsAppChecker struct {
	logger          domain.Logger
	installBasePath string
}

func NewWindowsAppChecker(logger domain.Logger, installBasePath string) *WindowsAppChecker {
	return &WindowsAppChecker{
		logger:          logger,
		installBasePath: installBasePath,
	}
}

func (c *WindowsAppChecker) IsInstalled(ctx context.Context, app *domain.App) (bool, error) {
	logger := c.logger.WithFields(map[string]interface{}{
		"app_id":   app.ID,
		"app_name": app.Name,
	})

	logger.Debug("Checking if app is installed")

	if app.CheckCmd != "" {
		if installed, err := c.checkRegistry(app.CheckCmd); err == nil && installed {
			logger.Debug("App found in registry")
			return true, nil
		}
	}

	if app.CheckPath != "" {
		// Use custom install path if available
		checkPath := app.CheckPath
		expandedPath := os.ExpandEnv(checkPath)
		if c.installBasePath != "" && !filepath.IsAbs(expandedPath) {
			expandedPath = filepath.Join(c.installBasePath, app.Name, expandedPath)
		}

		if installed, err := c.checkFilePath(expandedPath); err == nil && installed {
			logger.WithField("path", expandedPath).Debug("App found at specified path")
			return true, nil
		}
	}

	logger.Debug("App not found")
	return false, nil
}

func (c *WindowsAppChecker) checkRegistry(checkCmd string) (bool, error) {
	cmd := exec.Command("cmd", "/C", checkCmd)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (c *WindowsAppChecker) checkFilePath(checkPath string) (bool, error) {
	expandedPath := os.ExpandEnv(checkPath)
	
	if _, err := os.Stat(expandedPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check file path: %s", expandedPath)
	}
	
	return true, nil
}

func (c *WindowsAppChecker) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": getWindowsVersion(),
	}
}

func getWindowsVersion() string {
	cmd := exec.Command("cmd", "/C", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}
	return strings.TrimSpace(string(output))
} 
