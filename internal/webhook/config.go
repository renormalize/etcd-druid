// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/gardener/etcd-druid/internal/webhook/sentinel"
	"github.com/gardener/etcd-druid/internal/webhook/validate"

	flag "github.com/spf13/pflag"
)

// Config defines the configuration for etcd-druid webhooks.
type Config struct {
	// Sentinel is the configuration required for the sentinel webhook.
	Sentinel *sentinel.Config
	// Validate is the configuration required for the validating webhook.
	Validate *validate.Config
}

// InitFromFlags initializes the webhook config from the provided CLI flag set.
func (cfg *Config) InitFromFlags(fs *flag.FlagSet) {
	cfg.Sentinel = &sentinel.Config{}
	sentinel.InitFromFlags(fs, cfg.Sentinel)
	cfg.Validate = &validate.Config{}
	validate.InitFromFlags(fs, cfg.Validate)
}

// AtLeaseOneEnabled returns true if at least one webhook is enabled.
// NOTE for contributors: For every new webhook, add a disjunction condition with the webhook's Enabled field.
func (cfg *Config) AtLeaseOneEnabled() bool {
	return cfg.Sentinel.Enabled || cfg.Validate.Enabled
}
