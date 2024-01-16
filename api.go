package main

import (
	"errors"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/db"
	"gorm.io/gorm/clause"
)

// TODO: this should be moved to a separate object, not in Service

func (svc *Service) ListApps(userApps *[]App, apps *[]api.App) error {
	for _, app := range *userApps {
		apiApp := api.App{
			// ID:          app.ID,
			Name:        app.Name,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
			NostrPubkey: app.NostrPubkey,
		}

		var lastEvent NostrEvent
		result := svc.db.Where("app_id = ?", app.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if result.Error != nil {
			svc.Logger.Errorf("Failed to fetch last event %v", result.Error)
			return errors.New("Failed to fetch last event")
		}
		if result.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}
		*apps = append(*apps, apiApp)
	}
	return nil
}

func (svc *Service) GetInfo(info *api.InfoResponse) {
	info.BackendType = svc.cfg.LNBackendType
}

func (svc *Service) Setup(setupRequest *api.SetupRequest) error {
	dbConfigEntries := []db.ConfigEntry{}

	dbConfigEntries = append(dbConfigEntries, db.ConfigEntry{Key: "LN_BACKEND_TYPE", Value: setupRequest.LNBackendType})
	if setupRequest.BreezMnemonic != "" {
		dbConfigEntries = append(dbConfigEntries, db.ConfigEntry{Key: "BREEZ_MNEMONIC", Value: setupRequest.BreezMnemonic})
	}
	if setupRequest.GreenlightInviteCode != "" {
		dbConfigEntries = append(dbConfigEntries, db.ConfigEntry{Key: "GREENLIGHT_INVITE_CODE", Value: setupRequest.GreenlightInviteCode})
	}

	// Update columns to default value on `id` conflict
	res := svc.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&dbConfigEntries)

	if res.Error != nil {
		svc.Logger.Errorf("Failed to update config: %v", res.Error)
		return res.Error
	}

	return svc.launchLNBackend()
}
