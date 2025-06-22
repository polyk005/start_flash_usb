package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/polyk005/start_flash/internal/domain"
	"github.com/pkg/errors"
)

type InstallationService struct {
	appRepo           domain.AppRepository
	installationRepo  domain.InstallationRepository
	appChecker        domain.AppChecker
	appInstaller      domain.AppInstaller
	logger            domain.Logger
	installTimeout    time.Duration
	maxConcurrency    int
}

type InstallationServiceConfig struct {
	InstallTimeout time.Duration
	MaxConcurrency int
}

func NewInstallationService(
	appRepo domain.AppRepository,
	installationRepo domain.InstallationRepository,
	appChecker domain.AppChecker,
	appInstaller domain.AppInstaller,
	logger domain.Logger,
	config InstallationServiceConfig,
) *InstallationService {
	if config.InstallTimeout == 0 {
		config.InstallTimeout = 10 * time.Minute
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 3
	}

	return &InstallationService{
		appRepo:          appRepo,
		installationRepo: installationRepo,
		appChecker:       appChecker,
		appInstaller:     appInstaller,
		logger:           logger,
		installTimeout:   config.InstallTimeout,
		maxConcurrency:   config.MaxConcurrency,
	}
}

func (s *InstallationService) InstallAllApps(ctx context.Context) ([]*domain.InstallationResult, error) {
	s.logger.Info("Starting installation of all apps")

	apps, err := s.appRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get apps")
	}

	if len(apps) == 0 {
		s.logger.Warn("No apps found to install")
		return []*domain.InstallationResult{}, nil
	}

	s.logger.WithField("app_count", len(apps)).Info("Found apps to install")

	semaphore := make(chan struct{}, s.maxConcurrency)
	results := make([]*domain.InstallationResult, len(apps))
	var wg sync.WaitGroup

	for i, app := range apps {
		wg.Add(1)
		go func(index int, app *domain.App) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			result := s.installApp(ctx, app)
			results[index] = result
		}(i, app)
	}

	wg.Wait()

	s.logger.Info("All installations completed")
	return results, nil
}

func (s *InstallationService) InstallSelectedApps(ctx context.Context, appNames []string) ([]*domain.InstallationResult, error) {
	s.logger.WithField("apps", appNames).Info("Starting installation for selected apps")
	
	nameSet := make(map[string]struct{})
	for _, name := range appNames {
		nameSet[name] = struct{}{}
	}

	allApps, err := s.appRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get apps")
	}

	var appsToInstall []*domain.App
	for _, app := range allApps {
		if _, ok := nameSet[app.Name]; ok {
			appsToInstall = append(appsToInstall, app)
		}
	}

	if len(appsToInstall) == 0 {
		s.logger.Warn("No matching apps found to install")
		return []*domain.InstallationResult{}, nil
	}

	semaphore := make(chan struct{}, s.maxConcurrency)
	results := make([]*domain.InstallationResult, len(appsToInstall))
	var wg sync.WaitGroup

	for i, app := range appsToInstall {
		wg.Add(1)
		go func(index int, app *domain.App) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := s.installApp(ctx, app)
			results[index] = result
		}(i, app)
	}

	wg.Wait()
	
	s.logger.Info("Selected installations completed")
	return results, nil
}

func (s *InstallationService) InstallApp(ctx context.Context, appID string) (*domain.InstallationResult, error) {
	s.logger.WithField("app_id", appID).Info("Starting app installation")

	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get app %s", appID)
	}

	result := s.installApp(ctx, app)
	return result, nil
}

func (s *InstallationService) installApp(ctx context.Context, app *domain.App) *domain.InstallationResult {
	logger := s.logger.WithFields(map[string]interface{}{
		"app_id":   app.ID,
		"app_name": app.Name,
	})

	logger.Info("Processing app installation")

	isInstalled, err := s.appChecker.IsInstalled(ctx, app)
	if err != nil {
		logger.WithField("error", err).Error("Failed to check if app is installed")
		return &domain.InstallationResult{
			App:        app,
			IsInstalled: false,
			Error:      errors.Wrap(err, "failed to check installation status"),
		}
	}

	if isInstalled {
		logger.Info("App is already installed, skipping")
		return &domain.InstallationResult{
			App:        app,
			IsInstalled: true,
		}
	}

	installation := &domain.Installation{
		ID:        uuid.New().String(),
		AppID:     app.ID,
		Status:    domain.StatusInstalling,
		StartTime: time.Now(),
	}

	if err := s.installationRepo.Save(ctx, installation); err != nil {
		logger.WithField("error", err).Error("Failed to save installation record")
		return &domain.InstallationResult{
			App:        app,
			Installation: installation,
			Error:      errors.Wrap(err, "failed to save installation record"),
		}
	}

	installCtx, cancel := context.WithTimeout(ctx, s.installTimeout)
	defer cancel()

	_, err = s.appInstaller.Install(installCtx, app)
	if err != nil {
		installation.Status = domain.StatusFailed
		installation.Error = err.Error()
		installation.EndTime = &[]time.Time{time.Now()}[0]
		installation.Duration = installation.EndTime.Sub(installation.StartTime)

		if saveErr := s.installationRepo.Update(ctx, installation); saveErr != nil {
			logger.WithField("error", saveErr).Error("Failed to update failed installation record")
		}

		logger.WithField("error", err).Error("App installation failed")
		return &domain.InstallationResult{
			App:        app,
			Installation: installation,
			Error:      errors.Wrap(err, "installation failed"),
		}
	}

	installation.Status = domain.StatusSuccess
	installation.EndTime = &[]time.Time{time.Now()}[0]
	installation.Duration = installation.EndTime.Sub(installation.StartTime)

	if err := s.installationRepo.Update(ctx, installation); err != nil {
		logger.WithField("error", err).Error("Failed to update successful installation record")
	}

	logger.Info("App installation completed successfully")
	return &domain.InstallationResult{
		App:        app,
		Installation: installation,
		IsInstalled: true,
	}
}

func (s *InstallationService) GetInstallationHistory(ctx context.Context, appID string) ([]*domain.Installation, error) {
	s.logger.WithField("app_id", appID).Debug("Getting installation history")

	installations, err := s.installationRepo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get installation history for app %s", appID)
	}

	return installations, nil
}

func (s *InstallationService) GetInstallationStats(ctx context.Context) (*InstallationStats, error) {
	s.logger.Debug("Getting installation statistics")

	apps, err := s.appRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get apps for statistics")
	}

	stats := &InstallationStats{
		TotalApps: len(apps),
		Installed: 0,
		Failed:    0,
		Pending:   0,
	}

	for _, app := range apps {
		isInstalled, err := s.appChecker.IsInstalled(ctx, app)
		if err != nil {
			s.logger.WithField("app_id", app.ID).WithField("error", err).Warn("Failed to check app status")
			continue
		}

		if isInstalled {
			stats.Installed++
		} else {
			stats.Pending++
		}
	}

	return stats, nil
}

type InstallationStats struct {
	TotalApps int `json:"total_apps"`
	Installed int `json:"installed"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
}

func (s *InstallationStats) String() string {
	return fmt.Sprintf("Total: %d, Installed: %d, Failed: %d, Pending: %d",
		s.TotalApps, s.Installed, s.Failed, s.Pending)
}

// GetAllApps returns all applications from the repository.
func (s *InstallationService) GetAllApps(ctx context.Context) ([]*domain.App, error) {
	return s.appRepo.GetAll(ctx)
}

// IsAppInstalled checks if a specific app is installed.
func (s *InstallationService) IsAppInstalled(ctx context.Context, app *domain.App) (bool, error) {
	return s.appChecker.IsInstalled(ctx, app)
} 
