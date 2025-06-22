package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/polyk005/start_flash/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) GetAll(ctx context.Context) ([]*domain.App, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.App), args.Error(1)
}

func (m *MockAppRepository) GetByID(ctx context.Context, id string) (*domain.App, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.App), args.Error(1)
}

func (m *MockAppRepository) GetByName(ctx context.Context, name string) (*domain.App, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.App), args.Error(1)
}

func (m *MockAppRepository) Save(ctx context.Context, app *domain.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockInstallationRepository struct {
	mock.Mock
}

func (m *MockInstallationRepository) Save(ctx context.Context, installation *domain.Installation) error {
	args := m.Called(ctx, installation)
	return args.Error(0)
}

func (m *MockInstallationRepository) GetByAppID(ctx context.Context, appID string) ([]*domain.Installation, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).([]*domain.Installation), args.Error(1)
}

func (m *MockInstallationRepository) GetByID(ctx context.Context, id string) (*domain.Installation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Installation), args.Error(1)
}

func (m *MockInstallationRepository) Update(ctx context.Context, installation *domain.Installation) error {
	args := m.Called(ctx, installation)
	return args.Error(0)
}

type MockAppChecker struct {
	mock.Mock
}

func (m *MockAppChecker) IsInstalled(ctx context.Context, app *domain.App) (bool, error) {
	args := m.Called(ctx, app)
	return args.Bool(0), args.Error(1)
}

type MockAppInstaller struct {
	mock.Mock
}

func (m *MockAppInstaller) Install(ctx context.Context, app *domain.App) (*domain.Installation, error) {
	args := m.Called(ctx, app)
	return args.Get(0).(*domain.Installation), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) WithField(key string, value interface{}) domain.Logger {
	args := m.Called(key, value)
	return args.Get(0).(domain.Logger)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) domain.Logger {
	args := m.Called(fields)
	return args.Get(0).(domain.Logger)
}

func TestInstallationService_InstallAllApps(t *testing.T) {
	mockAppRepo := &MockAppRepository{}
	mockInstallationRepo := &MockInstallationRepository{}
	mockAppChecker := &MockAppChecker{}
	mockAppInstaller := &MockAppInstaller{}
	mockLogger := &MockLogger{}

	service := NewInstallationService(
		mockAppRepo,
		mockInstallationRepo,
		mockAppChecker,
		mockAppInstaller,
		mockLogger,
		InstallationServiceConfig{
			InstallTimeout: 5 * time.Minute,
			MaxConcurrency: 2,
		},
	)

	ctx := context.Background()

	apps := []*domain.App{
		{
			ID:   "app1",
			Name: "Test App 1",
			Path: "./test1.exe",
			Args: []string{"/silent"},
		},
		{
			ID:   "app2",
			Name: "Test App 2",
			Path: "./test2.exe",
			Args: []string{"/S"},
		},
	}

	mockAppRepo.On("GetAll", ctx).Return(apps, nil)
	mockLogger.On("Info", "Starting installation of all apps").Return()
	mockLogger.On("WithField", "app_count", 2).Return(mockLogger)
	mockLogger.On("Info", "Found apps to install").Return()
	mockLogger.On("Info", "All installations completed").Return()

	mockAppChecker.On("IsInstalled", ctx, apps[0]).Return(false, nil)
	mockAppChecker.On("IsInstalled", ctx, apps[1]).Return(false, nil)

	mockInstallationRepo.On("Save", ctx, mock.AnythingOfType("*domain.Installation")).Return(nil)
	mockInstallationRepo.On("Update", ctx, mock.AnythingOfType("*domain.Installation")).Return(nil)

	mockAppInstaller.On("Install", ctx, apps[0]).Return(&domain.Installation{
		ID:     "inst1",
		AppID:  "app1",
		Status: domain.StatusSuccess,
	}, nil)
	mockAppInstaller.On("Install", ctx, apps[1]).Return(&domain.Installation{
		ID:     "inst2",
		AppID:  "app2",
		Status: domain.StatusSuccess,
	}, nil)

	results, err := service.InstallAllApps(ctx)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].IsInstalled)
	assert.True(t, results[1].IsInstalled)
	assert.Nil(t, results[0].Error)
	assert.Nil(t, results[1].Error)

	mockAppRepo.AssertExpectations(t)
	mockInstallationRepo.AssertExpectations(t)
	mockAppChecker.AssertExpectations(t)
	mockAppInstaller.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestInstallationService_InstallApp(t *testing.T) {
	mockAppRepo := &MockAppRepository{}
	mockInstallationRepo := &MockInstallationRepository{}
	mockAppChecker := &MockAppChecker{}
	mockAppInstaller := &MockAppInstaller{}
	mockLogger := &MockLogger{}

	service := NewInstallationService(
		mockAppRepo,
		mockInstallationRepo,
		mockAppChecker,
		mockAppInstaller,
		mockLogger,
		InstallationServiceConfig{
			InstallTimeout: 5 * time.Minute,
			MaxConcurrency: 1,
		},
	)

	ctx := context.Background()
	appID := "test-app"
	app := &domain.App{
		ID:   appID,
		Name: "Test App",
		Path: "./test.exe",
		Args: []string{"/silent"},
	}

	mockAppRepo.On("GetByID", ctx, appID).Return(app, nil)
	mockLogger.On("WithField", "app_id", appID).Return(mockLogger)
	mockLogger.On("Info", "Starting app installation").Return()
	mockAppChecker.On("IsInstalled", ctx, app).Return(false, nil)
	mockInstallationRepo.On("Save", ctx, mock.AnythingOfType("*domain.Installation")).Return(nil)
	mockInstallationRepo.On("Update", ctx, mock.AnythingOfType("*domain.Installation")).Return(nil)
	mockAppInstaller.On("Install", ctx, app).Return(&domain.Installation{
		ID:     "inst1",
		AppID:  appID,
		Status: domain.StatusSuccess,
	}, nil)

	result, err := service.InstallApp(ctx, appID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsInstalled)
	assert.Nil(t, result.Error)
	assert.Equal(t, app, result.App)

	mockAppRepo.AssertExpectations(t)
	mockInstallationRepo.AssertExpectations(t)
	mockAppChecker.AssertExpectations(t)
	mockAppInstaller.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestInstallationService_GetInstallationStats(t *testing.T) {
	mockAppRepo := &MockAppRepository{}
	mockInstallationRepo := &MockInstallationRepository{}
	mockAppChecker := &MockAppChecker{}
	mockAppInstaller := &MockAppInstaller{}
	mockLogger := &MockLogger{}

	service := NewInstallationService(
		mockAppRepo,
		mockInstallationRepo,
		mockAppChecker,
		mockAppInstaller,
		mockLogger,
		InstallationServiceConfig{},
	)

	ctx := context.Background()
	apps := []*domain.App{
		{ID: "app1", Name: "App 1"},
		{ID: "app2", Name: "App 2"},
		{ID: "app3", Name: "App 3"},
	}

	mockAppRepo.On("GetAll", ctx).Return(apps, nil)
	mockLogger.On("Debug", "Getting installation statistics").Return()
	mockAppChecker.On("IsInstalled", ctx, apps[0]).Return(true, nil)
	mockAppChecker.On("IsInstalled", ctx, apps[1]).Return(false, nil)
	mockAppChecker.On("IsInstalled", ctx, apps[2]).Return(true, nil)

	stats, err := service.GetInstallationStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalApps)
	assert.Equal(t, 2, stats.Installed)
	assert.Equal(t, 1, stats.Pending)

	mockAppRepo.AssertExpectations(t)
	mockAppChecker.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
} 
