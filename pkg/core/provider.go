package core

import "github.com/supergiant/supergiant/pkg/models"

type Provider interface {
	ValidateAccount(*models.CloudAccount) error

	CreateKube(*models.Kube, *Action) error
	DeleteKube(*models.Kube) error

	CreateNode(*models.Node, *Action) error
	DeleteNode(*models.Node) error

	CreateVolume(*models.Volume, *Action) error
	WaitForVolumeAvailable(*models.Volume, *Action) error
	ResizeVolume(*models.Volume, *Action) error
	DeleteVolume(*models.Volume) error

	CreateEntrypoint(*models.Entrypoint, *Action) error
	AddPortToEntrypoint(*models.Entrypoint, int64, int64) error
	RemovePortFromEntrypoint(*models.Entrypoint, int64) error
	DeleteEntrypoint(*models.Entrypoint) error
}
