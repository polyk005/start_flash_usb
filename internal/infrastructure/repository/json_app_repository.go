package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/polyk005/start_flash/internal/domain"
	"github.com/pkg/errors"
)

type JSONAppRepository struct {
	filePath string
	logger   domain.Logger
}

type appData struct {
	Name        string   `json:"name"`
	URL         string   `json:"url,omitempty"`
	Path        string   `json:"path,omitempty"`
	Args        []string `json:"args"`
	CheckCmd    string   `json:"check,omitempty"`
	CheckPath   string   `json:"checkPath,omitempty"`
	OS          string   `json:"os"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
}

func NewJSONAppRepository(filePath string, logger domain.Logger) *JSONAppRepository {
	return &JSONAppRepository{
		filePath: filePath,
		logger:   logger,
	}
}

func (r *JSONAppRepository) GetAll(ctx context.Context) ([]*domain.App, error) {
	r.logger.Debug("Loading apps from JSON file", r.filePath)

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			r.logger.Warn("Apps configuration file not found, returning empty list")
			return []*domain.App{}, nil
		}
		return nil, errors.Wrapf(err, "failed to read apps file %s", r.filePath)
	}

	var appDataList []appData
	if err := json.Unmarshal(data, &appDataList); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal apps from %s", r.filePath)
	}

	apps := make([]*domain.App, len(appDataList))
	for i, data := range appDataList {
		app := r.convertToDomainApp(data)
		apps[i] = app
	}

	r.logger.WithField("app_count", len(apps)).Debug("Successfully loaded apps from JSON")
	return apps, nil
}

func (r *JSONAppRepository) GetByID(ctx context.Context, id string) (*domain.App, error) {
	apps, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.ID == id {
			return app, nil
		}
	}

	return nil, domain.ErrAppNotFound
}

func (r *JSONAppRepository) GetByName(ctx context.Context, name string) (*domain.App, error) {
	apps, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Name == name {
			return app, nil
		}
	}

	return nil, domain.ErrAppNotFound
}

func (r *JSONAppRepository) Save(ctx context.Context, app *domain.App) error {
	apps, err := r.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, existingApp := range apps {
		if existingApp.Name == app.Name && existingApp.ID != app.ID {
			return domain.ErrAppAlreadyExists
		}
	}

	if app.ID == "" {
		app.ID = uuid.New().String()
	}

	found := false
	for i, existingApp := range apps {
		if existingApp.ID == app.ID {
			apps[i] = app
			found = true
			break
		}
	}

	if !found {
		apps = append(apps, app)
	}

	return r.saveToFile(apps)
}

func (r *JSONAppRepository) Delete(ctx context.Context, id string) error {
	apps, err := r.GetAll(ctx)
	if err != nil {
		return err
	}

	found := false
	for i, app := range apps {
		if app.ID == id {
			apps = append(apps[:i], apps[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return domain.ErrAppNotFound
	}

	return r.saveToFile(apps)
}

func (r *JSONAppRepository) saveToFile(apps []*domain.App) error {
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dir)
	}

	appDataList := make([]appData, len(apps))
	for i, app := range apps {
		appDataList[i] = r.convertToAppData(app)
	}

	data, err := json.MarshalIndent(appDataList, "", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal apps to JSON")
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return errors.Wrapf(err, "failed to write apps to file %s", r.filePath)
	}

	r.logger.WithField("app_count", len(apps)).Debug("Successfully saved apps to JSON file")
	return nil
}

func (r *JSONAppRepository) convertToDomainApp(data appData) *domain.App {
	id := uuid.NewSHA1(uuid.Nil, []byte(fmt.Sprintf("%s-%s", data.Name, data.OS))).String()

	return &domain.App{
		ID:          id,
		Name:        data.Name,
		URL:         data.URL,
		Path:        data.Path,
		Args:        data.Args,
		CheckCmd:    data.CheckCmd,
		CheckPath:   data.CheckPath,
		OS:          data.OS,
		Version:     data.Version,
		Description: data.Description,
		Category:    data.Category,
	}
}

func (r *JSONAppRepository) convertToAppData(app *domain.App) appData {
	return appData{
		Name:        app.Name,
		URL:         app.URL,
		Path:        app.Path,
		Args:        app.Args,
		CheckCmd:    app.CheckCmd,
		CheckPath:   app.CheckPath,
		OS:          app.OS,
		Version:     app.Version,
		Description: app.Description,
		Category:    app.Category,
	}
} 
