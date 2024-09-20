package server_registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/itimofeev/yas3/internal/entity"
	"github.com/itimofeev/yas3/internal/provider/store"
)

type Config struct {
	StoreServerAddrs []string `validate:"required"`
}

// Registry stores information about store servers. Periodically checks store servers available space in order to use the least loaded servers first.
type Registry struct {
	storeClients map[string]entity.StoreClient

	mostFreeClients []entity.StoreClient
	states          map[string]StoreServerState
	muState         sync.RWMutex
}

func New(ctx context.Context, cfg Config) (*Registry, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	storeClients := make(map[string]entity.StoreClient)
	for _, storeAddr := range cfg.StoreServerAddrs {
		client, err := store.New(store.Config{
			StoreAddr: storeAddr,
		})
		if err != nil {
			return nil, err
		}
		storeClients[client.GetID()] = client
	}

	r := &Registry{
		storeClients: storeClients,
	}
	r.updateStates(ctx)
	return r, nil
}

// GetServersForParts returns the least loaded servers to store file parts. Uses server states to decide which servers are more free.
func (r *Registry) GetServersForParts(nFileParts int64) ([]entity.StoreClient, error) {
	r.muState.RLock()
	defer r.muState.RUnlock()

	if len(r.mostFreeClients) == 0 {
		return nil, errors.New("all stores are offline")
	}

	storeClients := make([]entity.StoreClient, 0, nFileParts)
	for n := range nFileParts {
		storeClients = append(storeClients, r.mostFreeClients[n%int64(len(r.mostFreeClients))])
	}
	return storeClients, nil
}

// GetStoreClients returns list of clients to store servers that can be used to download files from.
func (r *Registry) GetStoreClients(serverIDs []string) ([]entity.StoreClient, error) {
	r.muState.RLock()
	defer r.muState.RUnlock()

	clients := make([]entity.StoreClient, 0, len(serverIDs))
	for _, serverID := range serverIDs {
		state := r.states[serverID]
		if !state.IsOnline {
			return nil, fmt.Errorf("storeClient for server %s is offline", serverID)
		}
		clients = append(clients, r.storeClients[serverID])
	}
	return clients, nil
}

// Run periodically asks store servers about their space statistics.
func (r *Registry) Run(ctx context.Context) error {
	t := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-t.C:
			r.updateStates(ctx)
			t.Reset(time.Second * 10)
		case <-ctx.Done():
			return nil
		}
	}
}

func (r *Registry) receiveNewStates(ctx context.Context) map[string]StoreServerState {
	states := make(map[string]StoreServerState)
	for _, client := range r.storeClients {
		space, err := client.GetAvailableSpace(ctx)
		if err != nil {
			slog.Warn("store client returned error", "id", client.GetID(), "err", err)
			states[client.GetID()] = StoreServerState{
				ID:       client.GetID(),
				IsOnline: false,
			}
			continue
		}
		states[client.GetID()] = StoreServerState{
			ID:       client.GetID(),
			Space:    space,
			IsOnline: true,
		}
	}
	return states
}

func (r *Registry) updateStates(ctx context.Context) {
	newStates := r.receiveNewStates(ctx)

	r.muState.Lock()
	defer r.muState.Unlock()

	newFreeClients := make([]entity.StoreClient, 0, len(r.storeClients))
	for _, state := range newStates {
		if state.IsOnline {
			newFreeClients = append(newFreeClients, r.storeClients[state.ID])
		}
	}
	slices.SortFunc(newFreeClients, func(a, b entity.StoreClient) int {
		return newStates[a.GetID()].GetAvailableSpacePercent() - newStates[b.GetID()].GetAvailableSpacePercent()
	})

	r.states = newStates
	r.mostFreeClients = newFreeClients
	slog.Debug("new store server states received", "states", newStates, "serversOnline", len(newFreeClients))
}

type StoreServerState struct {
	ID       string
	Space    entity.AvailableSpace
	IsOnline bool
}

func (s StoreServerState) GetAvailableSpacePercent() int {
	return int(float64(s.Space.Used) / float64(s.Space.Total))
}
