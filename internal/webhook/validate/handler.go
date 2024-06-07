// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Might not be necessary since CONNECT isn't defined in the ValidatingWebhookConfiguration
var allowedOperations = []admissionv1.Operation{admissionv1.Connect}

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
// Handle should handle DELETE, CREATE, UPDATE in that order
func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.logger.Info("The validating webhook was invoked")
	h.logger.Info(fmt.Sprintf("The request was: %s", req.Operation))
	requestGKString := fmt.Sprintf("%s/%s", req.Kind.Group, req.Kind.Kind)
	log := h.logger.WithValues("name", req.Name, "namespace", req.Namespace, "resourceGroupKind", requestGKString, "operation", req.Operation, "user", req.UserInfo.Username)
	log.V(1).Info("Validating webhook invoked")

	// Might not be necessary since CONNECT isn't defined in the ValidatingWebhookConfiguration
	if slices.Contains(allowedOperations, req.Operation) {
		h.logger.Info("The validating webhook was called for CONNECT operation")
		return admission.Allowed(fmt.Sprintf("operation %s is allowed", req.Operation))
	}

	// etcd, err := h.getRelevantEtcdForRequest(req)
	oldEtcd, newEtcd := &druidv1alpha1.Etcd{}, &druidv1alpha1.Etcd{}
	err := h.DecodeRequestObjects(req, oldEtcd, newEtcd)
	if err != nil {
		h.logger.Info("The validating webhook errored")
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("operation %s is denied due to error: %w", req.Operation, err))
	}
	fmt.Println("the old object are: ", oldEtcd)
	fmt.Println("the new object are: ", newEtcd)

	// etcd deletion request should be denied if it is already in deletion or is in reconciliation
	if req.Operation == admissionv1.Delete && (oldEtcd.IsReconciliationInProgress() || oldEtcd.IsDeletionInProgress()) {
		var message string
		if oldEtcd.IsDeletionInProgress() {
			message = "etcd deletion in progress"
		} else {
			message = "etcd reconciliation in progress"
		}
		h.logger.Info("The validating webhook denies")
		return admission.Denied(fmt.Sprintf("operation %s is denied: %s", req.Operation, message))
	}

	h.logger.Info("The validating webhook approves")
	fmt.Println("Name of the newEtcd is: ", newEtcd.Name)
	return admission.Allowed(fmt.Sprintf("operation %s is allowed", req.Operation))
}

// getRelevantEtcdForRequest returns the Etcd resource that is relevant to the request.
// CREATE would only require information about `Object` since only that will be present.
// UPDATE would present information about `OldObject` and `Object`, but we can validate the Etcd by using `Object` only.
// DELETE would only require information about the `OldObject`, since `Object` would be empty.
// func (h *Handler) getRelevantEtcdForRequest(req admission.Request) (*druidv1alpha1.Etcd, error) {
// 	if req.Operation == admissionv1.Delete {
// 		oldEtcd := &druidv1alpha1.Etcd{}
// 		if err := h.decoder.DecodeRaw(req.OldObject, oldEtcd); err != nil {
// 			return nil, err
// 		}
// 		return oldEtcd, nil
// 	}
// 	etcd := &druidv1alpha1.Etcd{}
// 	if err := h.decoder.Decode(req, etcd); err != nil {
// 		return nil, err
// 	}
// 	return etcd, nil
// }

// DecodeRequestObjects decodes the runtime.RawExtension req.Object and req.OldObject into the passed oldObj and newObj's concrete type
func (h *Handler) DecodeRequestObjects(req admission.Request, oldObj runtime.Object, newObj runtime.Object) error {
	if req.Operation == admissionv1.Delete {
		if err := h.decoder.DecodeRaw(req.OldObject, oldObj); err != nil {
			return err
		}
		// // TODO: @renormalize UPDATE would require both oldObj and newObj, so the following return has to be removed
		// return nil
	}
	// there is no new object
	if len(req.Object.Raw) == 0 {
		return nil
	}
	if err := h.decoder.Decode(req, newObj); err != nil {
		return err
	}
	return nil
}

// func (h *Handler) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
// 	object := obj.(*druidv1alpha1.Etcd)
// 	if errs := etcdvalidation.ValidateEtcd(object); len(errs) > 0 {
// 		return nil, apierrors.NewInvalid(object.GroupVersionKind().GroupKind(), object.GetName(), errs)
// 	}
// 	h.logger.Info("The validating webhook was called for create operation")
// 	return nil, nil
// }

// func (h *Handler) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
// 	object := newObj.(*druidv1alpha1.Etcd)
// 	if errs := etcdvalidation.ValidateEtcdUpdate(object, oldObj.(*druidv1alpha1.Etcd)); len(errs) > 0 {
// 		return nil, apierrors.NewInvalid(object.GroupVersionKind().GroupKind(), object.GetName(), errs)
// 	}
// 	h.logger.Info("The validating webhook was called for update operation")
// 	return nil, nil
// }

// func (h *Handler) ValidateDelete(_ context.Context, newObj runtime.Object) (admission.Warnings, error) {
// 	object := newObj.(*druidv1alpha1.Etcd)
// 	if errs := etcdvalidation.ValidateEtcdUpdate(object, newObj.(*druidv1alpha1.Etcd)); len(errs) > 0 {
// 		return nil, apierrors.NewInvalid(object.GroupVersionKind().GroupKind(), object.GetName(), errs)
// 	}
// 	h.logger.Info("The validating webhook was called for update operation")
// 	return nil, nil
// }
