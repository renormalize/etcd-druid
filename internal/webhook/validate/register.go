// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// handlerName is the name of the webhook handler.
	handlerName = "validating-webhook"
	// WebhookPath is the path at which the handler should be registered.
	webhookPath = "/webhooks/validate"
)

// RegisterWithManager registers Handler to the given manager.
func (h *Handler) RegisterWithManager(mgr manager.Manager) error {
	webhook := &admission.Webhook{
		Handler:      h,
		RecoverPanic: true,
	}
	mgr.GetWebhookServer().Register(webhookPath, webhook)
	return nil
}
