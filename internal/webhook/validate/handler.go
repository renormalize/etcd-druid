// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"fmt"

	"github.com/gardener/etcd-druid/api/v1alpha1"
	druidvalidation "github.com/gardener/etcd-druid/api/validation"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

// NewHandler creates a new handler for Validating Webhook.
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

func (h *Handler) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	object := obj.(*v1alpha1.Etcd)
	if errs := druidvalidation.ValidateEtcd(object); len(errs) > 0 {
		return nil, apierrors.NewInvalid(object.GroupVersionKind().GroupKind(), object.GetName(), errs)
	}
	h.logger.Info("The validating webhook was called for create operation")
	return nil, nil
}

func (h *Handler) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	object := newObj.(*v1alpha1.Etcd)
	if errs := druidvalidation.ValidateEtcdUpdate(object, oldObj.(*v1alpha1.Etcd)); len(errs) > 0 {
		return nil, apierrors.NewInvalid(object.GroupVersionKind().GroupKind(), object.GetName(), errs)
	}
	h.logger.Info("The validating webhook was called for update operation")
	return nil, nil
}

func (h *Handler) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	h.logger.Info("The validating webhook was called for delete operation")
	return nil, nil
}
