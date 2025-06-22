package repository

import (
	"context"
	"sync"
	"time"

	"github.com/polyk005/start_flash/internal/domain"
	"github.com/pkg/errors"
)

type MemoryInstallationRepository struct {
	installations map[string]*domain.Installation
	appInstallations map[string][]*domain.Installation // appID -> installations
	mutex         sync.RWMutex
	logger        domain.Logger
}

func NewMemoryInstallationRepository(logger domain.Logger) *MemoryInstallationRepository {
	return &MemoryInstallationRepository{
		installations:    make(map[string]*domain.Installation),
		appInstallations: make(map[string][]*domain.Installation),
		logger:           logger,
	}
}

func (r *MemoryInstallationRepository) Save(ctx context.Context, installation *domain.Installation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if installation.ID == "" {
		return errors.New("installation ID cannot be empty")
	}

	r.installations[installation.ID] = installation

	if _, exists := r.appInstallations[installation.AppID]; !exists {
		r.appInstallations[installation.AppID] = make([]*domain.Installation, 0)
	}
	r.appInstallations[installation.AppID] = append(r.appInstallations[installation.AppID], installation)

	r.logger.WithFields(map[string]interface{}{
		"installation_id": installation.ID,
		"app_id":          installation.AppID,
		"status":          installation.Status,
	}).Debug("Saved installation record")

	return nil
}

func (r *MemoryInstallationRepository) GetByAppID(ctx context.Context, appID string) ([]*domain.Installation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	installations, exists := r.appInstallations[appID]
	if !exists {
		return []*domain.Installation{}, nil
	}

	result := make([]*domain.Installation, len(installations))
	copy(result, installations)

	r.logger.WithField("app_id", appID).WithField("count", len(result)).Debug("Retrieved installations for app")
	return result, nil
}

func (r *MemoryInstallationRepository) GetByID(ctx context.Context, id string) (*domain.Installation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	installation, exists := r.installations[id]
	if !exists {
		return nil, errors.New("installation not found")
	}

	r.logger.WithField("installation_id", id).Debug("Retrieved installation by ID")
	return installation, nil
}

func (r *MemoryInstallationRepository) Update(ctx context.Context, installation *domain.Installation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if installation.ID == "" {
		return errors.New("installation ID cannot be empty")
	}

	if _, exists := r.installations[installation.ID]; !exists {
		return errors.New("installation not found")
	}

	r.installations[installation.ID] = installation

	if appInstallations, exists := r.appInstallations[installation.AppID]; exists {
		for i, existingInstallation := range appInstallations {
			if existingInstallation.ID == installation.ID {
				appInstallations[i] = installation
				break
			}
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"installation_id": installation.ID,
		"app_id":          installation.AppID,
		"status":          installation.Status,
	}).Debug("Updated installation record")

	return nil
}

func (r *MemoryInstallationRepository) GetRecentInstallations(ctx context.Context, limit int) ([]*domain.Installation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	allInstallations := make([]*domain.Installation, 0, len(r.installations))
	for _, installation := range r.installations {
		allInstallations = append(allInstallations, installation)
	}

	sortInstallationsByTime(allInstallations)

	if len(allInstallations) > limit {
		allInstallations = allInstallations[:limit]
	}

	r.logger.WithField("count", len(allInstallations)).Debug("Retrieved recent installations")
	return allInstallations, nil
}

func (r *MemoryInstallationRepository) GetInstallationStats(ctx context.Context) (*InstallationStats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := &InstallationStats{
		Total:      0,
		Successful: 0,
		Failed:     0,
		Pending:    0,
		Installing: 0,
	}

	for _, installation := range r.installations {
		stats.Total++
		switch installation.Status {
		case domain.StatusSuccess:
			stats.Successful++
		case domain.StatusFailed:
			stats.Failed++
		case domain.StatusPending:
			stats.Pending++
		case domain.StatusInstalling:
			stats.Installing++
		}
	}

	return stats, nil
}

func (r *MemoryInstallationRepository) CleanupOldInstallations(ctx context.Context, olderThan time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cutoffTime := time.Now().Add(-olderThan)
	removedCount := 0

	toRemove := make([]string, 0)
	for id, installation := range r.installations {
		if installation.StartTime.Before(cutoffTime) {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		installation := r.installations[id]
		
		delete(r.installations, id)
		
		if appInstallations, exists := r.appInstallations[installation.AppID]; exists {
			for i, existingInstallation := range appInstallations {
				if existingInstallation.ID == id {
					r.appInstallations[installation.AppID] = append(appInstallations[:i], appInstallations[i+1:]...)
					break
				}
			}
		}
		
		removedCount++
	}

	r.logger.WithField("removed_count", removedCount).Info("Cleaned up old installations")
	return nil
}

type InstallationStats struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
	Pending    int `json:"pending"`
	Installing int `json:"installing"`
}

func sortInstallationsByTime(installations []*domain.Installation) {
	for i := 0; i < len(installations)-1; i++ {
		for j := i + 1; j < len(installations); j++ {
			if installations[i].StartTime.Before(installations[j].StartTime) {
				installations[i], installations[j] = installations[j], installations[i]
			}
		}
	}
} 
