// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/gardener/etcd-druid/internal/webhook/sentinel"
	"github.com/gardener/etcd-druid/internal/webhook/validate"

	"golang.org/x/exp/slog"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Register registers all etcd-druid webhooks with the controller manager.
func Register(mgr ctrl.Manager, config *Config) error {
	// Add sentinel webhook to the manager
	if config.Sentinel.Enabled {
		sentinelWebhook, err := sentinel.NewHandler(
			mgr,
			config.Sentinel,
		)
		if err != nil {
			return err
		}
		slog.Info("Registering Sentinel Webhook with manager")
		if err := sentinelWebhook.RegisterWithManager(mgr); err != nil {
			return err
		}
	}
	if config.Validate.Enabled {
		validatingWebhook, err := validate.NewHandler(
			mgr,
			config.Validate,
		)
		if err != nil {
			return err
		}
		slog.Info("Registering Validating Webhook with manager")
		if err := validatingWebhook.RegisterWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
