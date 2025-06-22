package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/polyk005/start_flash/internal/config"
	"github.com/polyk005/start_flash/internal/domain"
	"github.com/polyk005/start_flash/internal/infrastructure/installer"
	"github.com/polyk005/start_flash/internal/infrastructure/logger"
	"github.com/polyk005/start_flash/internal/infrastructure/repository"
	"github.com/polyk005/start_flash/internal/usecase"
)

func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

func isInputFromTerminal() bool {
	fileInfo, _ := os.Stdin.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func main() {
	// Add a small delay to keep the window open in case of immediate error
	defer func() {
		if !isInputFromTerminal() {
			time.Sleep(5 * time.Second)
		}
	}()

	// Get the directory where the executable is located
	exeDir := getExecutableDir()
	configPath := filepath.Join(exeDir, "configs", "config.yaml")

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		if isInputFromTerminal() {
			fmt.Println("\nPress Enter to exit...")
			fmt.Scanln()
		}
		os.Exit(1)
	}

	// Update paths to be relative to executable directory
	cfg.Install.AppsConfigPath = filepath.Join(exeDir, "configs", "apps.json")
	cfg.Install.AppsDirectory = filepath.Join(exeDir, "apps")
	if cfg.Install.InstallBasePath != "" {
		cfg.Install.InstallBasePath = filepath.Join(exeDir, cfg.Install.InstallBasePath)
	}

	log := logger.NewLogrusLogger(cfg.Logging.Level, cfg.Logging.Format)
	log.Info("Starting StartFlash application")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Received signal, shutting down.")
		cancel()
	}()

	appRepo := repository.NewJSONAppRepository(cfg.Install.AppsConfigPath, log)
	installationRepo := repository.NewMemoryInstallationRepository(log)
	appChecker := installer.NewWindowsAppChecker(log, cfg.Install.InstallBasePath)
	appInstaller := installer.NewWindowsInstaller(log, cfg.Install.InstallBasePath)

	installationService := usecase.NewInstallationService(
		appRepo,
		installationRepo,
		appChecker,
		appInstaller,
		log,
		usecase.InstallationServiceConfig{
			InstallTimeout: cfg.Install.Timeout,
			MaxConcurrency: cfg.Install.MaxConcurrency,
		},
	)

	if err := run(ctx, installationService, log); err != nil {
		log.Error("Application failed", "error", err)
		if isInputFromTerminal() {
			fmt.Println("\nPress Enter to exit...")
			fmt.Scanln()
		}
		os.Exit(1)
	}

	log.Info("Application completed successfully")

	if isInputFromTerminal() {
		fmt.Println("\nPress Enter to exit...")
		fmt.Scanln()
	}
}

func run(ctx context.Context, service *usecase.InstallationService, log domain.Logger) error {
	showStartupAnimation()

	allApps, err := service.GetAllApps(ctx)
	if err != nil {
		return fmt.Errorf("failed to get application list: %w", err)
	}

	var options []string
	appMap := make(map[string]*domain.App)
	for _, app := range allApps {
		isInstalled, _ := service.IsAppInstalled(ctx, app)
		status := ""
		if isInstalled {
			status = " (✅ installed)"
		}
		option := fmt.Sprintf("%s%s", app.Name, status)
		options = append(options, option)
		appMap[option] = app
	}

	var selectedOptions []string
	prompt := &survey.MultiSelect{
		Message: "Select applications to install (use spacebar, then enter):",
		Options: options,
		PageSize: 15,
	}
	survey.AskOne(prompt, &selectedOptions)

	if len(selectedOptions) == 0 {
		log.Info("No applications selected. Exiting.")
		return nil
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("               STARTING INSTALLATION")
	fmt.Println(strings.Repeat("=", 60))

	var results []*domain.InstallationResult
	for _, option := range selectedOptions {
		app := appMap[option]
		
		fmt.Printf("▶️  Processing %s...\n", app.Name)

		result, err := service.InstallApp(ctx, app.ID)
		if err != nil {
			result = &domain.InstallationResult{App: app, Error: err}
		}
		results = append(results, result)
		
		if result.Error != nil {
			fmt.Printf("❌ %-20s | FAILED\n", result.App.Name)
			log.Error("Installation failed", "app", result.App.Name, "error", result.Error)
		} else if result.IsInstalled {
			if result.Installation != nil {
				duration := result.Installation.Duration.Round(time.Second)
				if duration > 0 {
					fmt.Printf("✅ %-20s | SUCCESS (%s)\n", result.App.Name, duration)
				} else {
					fmt.Printf("✅ %-20s | SUCCESS\n", result.App.Name)
				}
			} else {
				fmt.Printf("✅ %-20s | ALREADY INSTALLED\n", result.App.Name)
			}
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	showSummary(results)
	return nil
}

func showSummary(results []*domain.InstallationResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    INSTALLATION SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	successCount := 0
	failedCount := 0
	
	for _, result := range results {
		if result.Error != nil {
			failedCount++
		} else {
			successCount++
		}
	}
	
	fmt.Printf("Total Selected: %d | Success: %d | Failed: %d\n",
		len(results), successCount, failedCount)
	fmt.Println(strings.Repeat("=", 60))
}

func showStartupAnimation() {
	fmt.Print(`
  _____ _______   _    _  _____ ______  
 / ____|__   __| | |  | ||  __ \|  ____| 
| (___    | |    | |  | || |__) | |__    
 \___ \   | |    | |  | ||  _  /|  __|   
 ____) |  | |    | |__| || | \ \| |____  
|_____/   |_|     \____/ |_|  \_\______| 
                                          
    StartFlash - Enterprise Edition
`)
	time.Sleep(1 * time.Second)
} 