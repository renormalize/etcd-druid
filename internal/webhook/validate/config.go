package validate

import (
	flag "github.com/spf13/pflag"
)

const (
	enableValidatingWebhookFlagName = "enable-validating-webhook"
	defaultEnableValidatingWebhook  = true
)

// Config defines the configuration for the Validating Webhook.
type Config struct {
	// Enabled indicates whether the Validating Webhook is enabled.
	Enabled bool
}

// InitFromFlags initializes the config from the provided CLI flag set.
func InitFromFlags(fs *flag.FlagSet, cfg *Config) {
	fs.BoolVar(&cfg.Enabled, enableValidatingWebhookFlagName, defaultEnableValidatingWebhook,
		"Enable Validating Webhook to valdiate the creation of Etcd objects.")
}
