package podman

import (
	"context"
	"fmt"
	"strings"
)

// VolumeManager handles podman volumes for database engines.
type VolumeManager struct {
	client *PodmanClient
}

// NewVolumeManager creates a new VolumeManager.
func NewVolumeManager(client *PodmanClient) *VolumeManager {
	return &VolumeManager{client: client}
}

// Ensure creates a volume with the given name if it doesn't exist.
func (vm *VolumeManager) Ensure(ctx context.Context, name string) error {
	exists, err := vm.client.VolumeExists(ctx, name)
	if err != nil {
		return fmt.Errorf("check volume: %w", err)
	}
	if exists {
		return nil
	}

	_, err = vm.client.Run(ctx, "volume", "create", name)
	if err != nil {
		return fmt.Errorf("create volume: %w", err)
	}
	return nil
}

// Remove deletes a volume.
func (vm *VolumeManager) Remove(ctx context.Context, name string) error {
	exists, err := vm.client.VolumeExists(ctx, name)
	if err != nil {
		return fmt.Errorf("check volume: %w", err)
	}
	if !exists {
		return nil
	}

	_, err = vm.client.Run(ctx, "volume", "rm", name)
	if err != nil {
		return fmt.Errorf("remove volume: %w", err)
	}
	return nil
}

// Exists returns true if the volume exists.
func (vm *VolumeManager) Exists(ctx context.Context, name string) (bool, error) {
	return vm.client.VolumeExists(ctx, name)
}

// Inspect returns volume details.
func (vm *VolumeManager) Inspect(ctx context.Context, name string) (string, error) {
	output, err := vm.client.Run(ctx, "volume", "inspect", name)
	if err != nil {
		if strings.Contains(err.Error(), "volume does not exist") {
			return "", ErrVolumeNotFound
		}
		return "", err
	}
	return output, nil
}
