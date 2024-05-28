// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Handler is the Validating Webhook admission handler.
type Handler struct {
	client.Client
	config  *Config
	decoder *admission.Decoder
	logger  logr.Logger
}

// NewHandler creates a new handler for Sentinel Webhook.
func NewHandler(mgr manager.Manager, config *Config) (*Handler, error) {
	decoder := admission.NewDecoder(mgr.GetScheme())
	return &Handler{
		Client:  mgr.GetClient(),
		config:  config,
		decoder: decoder,
		logger:  mgr.GetLogger().WithName(handlerName),
	}, nil
}

// Handle handles admission requests and validates creation of Etcd resources.
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Println("The validating webhook was called")
	h.logger.Info("The validating webhook was called")
	return admission.Allowed(fmt.Sprintf("valid etcd-druid"))
}
