package domain

import (
	"context"
	"time"
)

type App struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Path        string   `json:"path"`
	Args        []string `json:"args"`
	InstallDirArgTpl string   `json:"installDirArgTpl"`
	CheckCmd    string   `json:"check_cmd"`
	CheckPath   string   `json:"check_path"`
	OS          string   `json:"os"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category"`
}

type InstallationStatus string

const (
	StatusPending   InstallationStatus = "pending"
	StatusInstalling InstallationStatus = "installing"
	StatusSuccess   InstallationStatus = "success"
	StatusFailed    InstallationStatus = "failed"
	StatusSkipped   InstallationStatus = "skipped"
)

type Installation struct {
	ID        string            `json:"id"`
	AppID     string            `json:"app_id"`
	Status    InstallationStatus `json:"status"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration,omitempty"`
	Error     string            `json:"error,omitempty"`
	Logs      []string          `json:"logs,omitempty"`
}

type InstallationResult struct {
	App        *App
	Installation *Installation
	IsInstalled bool
	Error      error
}

type AppRepository interface {
	GetAll(ctx context.Context) ([]*App, error)
	GetByID(ctx context.Context, id string) (*App, error)
	GetByName(ctx context.Context, name string) (*App, error)
	Save(ctx context.Context, app *App) error
	Delete(ctx context.Context, id string) error
}

type InstallationRepository interface {
	Save(ctx context.Context, installation *Installation) error
	GetByAppID(ctx context.Context, appID string) ([]*Installation, error)
	GetByID(ctx context.Context, id string) (*Installation, error)
	Update(ctx context.Context, installation *Installation) error
}

type AppChecker interface {
	IsInstalled(ctx context.Context, app *App) (bool, error)
}

type AppInstaller interface {
	Install(ctx context.Context, app *App) (*Installation, error)
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
} 
