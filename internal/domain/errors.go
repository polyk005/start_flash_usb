package domain

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	ErrAppNotFound      = errors.New("app not found")
	ErrAppAlreadyExists = errors.New("app already exists")
	ErrInvalidApp       = errors.New("invalid app configuration")
	ErrInstallationFailed = errors.New("installation failed")
	ErrAppNotInstalled  = errors.New("app not installed")
	ErrTimeout          = errors.New("operation timeout")
	ErrFileNotFound     = errors.New("file not found")
	ErrPermissionDenied = errors.New("permission denied")
)

type AppError struct {
	AppID string
	Op    string
	Err   error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("app %s: %s: %v", e.AppID, e.Op, e.Err)
	}
	return fmt.Sprintf("app %s: %s", e.AppID, e.Op)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

type InstallationError struct {
	InstallationID string
	AppID          string
	Op             string
	Err            error
}

func (e *InstallationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("installation %s (app %s): %s: %v", 
			e.InstallationID, e.AppID, e.Op, e.Err)
	}
	return fmt.Sprintf("installation %s (app %s): %s", 
		e.InstallationID, e.AppID, e.Op)
}

func (e *InstallationError) Unwrap() error {
	return e.Err
}

func NewAppError(appID, op string, err error) *AppError {
	return &AppError{
		AppID: appID,
		Op:    op,
		Err:   err,
	}
}

func NewInstallationError(installationID, appID, op string, err error) *InstallationError {
	return &InstallationError{
		InstallationID: installationID,
		AppID:          appID,
		Op:             op,
		Err:            err,
	}
} 
