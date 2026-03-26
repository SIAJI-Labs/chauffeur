package podman

import (
	"context"
	"fmt"
	"strings"
)

// NetworkManager handles the chauf-net network.
type NetworkManager struct {
	client *PodmanClient
}

// NewNetworkManager creates a new NetworkManager.
func NewNetworkManager(client *PodmanClient) *NetworkManager {
	return &NetworkManager{client: client}
}

// Ensure creates the chauf-net network if it doesn't exist.
func (nm *NetworkManager) Ensure(ctx context.Context) error {
	exists, err := nm.client.NetworkExists(ctx, NetworkName)
	if err != nil {
		return fmt.Errorf("check network: %w", err)
	}
	if exists {
		return nil
	}

	// Create the network
	_, err = nm.client.Run(ctx,
		"network", "create",
		"--driver", "bridge",
		NetworkName,
	)
	if err != nil {
		return fmt.Errorf("create network: %w", err)
	}
	return nil
}

// Remove deletes the chauf-net network.
func (nm *NetworkManager) Remove(ctx context.Context) error {
	exists, err := nm.client.NetworkExists(ctx, NetworkName)
	if err != nil {
		return fmt.Errorf("check network: %w", err)
	}
	if !exists {
		return nil
	}

	_, err = nm.client.Run(ctx, "network", "rm", NetworkName)
	if err != nil {
		return fmt.Errorf("remove network: %w", err)
	}
	return nil
}

// Exists returns true if the network exists.
func (nm *NetworkManager) Exists(ctx context.Context) (bool, error) {
	return nm.client.NetworkExists(ctx, NetworkName)
}

// Inspect returns network details.
func (nm *NetworkManager) Inspect(ctx context.Context) (string, error) {
	output, err := nm.client.Run(ctx, "network", "inspect", NetworkName)
	if err != nil {
		if strings.Contains(err.Error(), "network not found") {
			return "", ErrNetworkNotFound
		}
		return "", err
	}
	return output, nil
}
