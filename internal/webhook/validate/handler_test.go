// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/gardener/etcd-druid/internal/client/kubernetes"

	testutils "github.com/gardener/etcd-druid/test/utils"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	etcdGVK = metav1.GroupVersionKind{Group: "druid.gardener.cloud", Version: "v1alpha1", Kind: "Etcd"}
)

// Probably not necessary since the webhook configuration does not specify CONNECT
func TestHandleConnect(t *testing.T) {
	g := NewWithT(t)

	testCases := []struct {
		name        string
		operation   admissionv1.Operation
		expectedMsg string
	}{
		{
			name:        "allow connect operation for the etcd resource",
			operation:   admissionv1.Connect,
			expectedMsg: "operation CONNECT is allowed",
		},
	}
	cl := testutils.CreateDefaultFakeClient()
	decoder := admission.NewDecoder(cl.Scheme())

	handler := &Handler{
		Client: cl,
		config: &Config{
			Enabled: true,
		},
		decoder: decoder,
		logger:  logr.Discard(),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := handler.Handle(context.Background(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: tc.operation,
					// Connect for the Etcd resource is allowed.
					Kind: etcdGVK,
				},
			})
			g.Expect(resp.Allowed).To(BeTrue())
			g.Expect(resp.Result.Message).To(Equal(tc.expectedMsg))
		})
	}
}

func TestHandleDelete(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("aasdff ", r)
		}
	}()
	g := NewWithT(t)

	testCases := []struct {
		name string
		// request
		operation          admissionv1.Operation
		lastOperationType  druidv1alpha1.LastOperationType
		lastOperationState druidv1alpha1.LastOperationState
		// expected
		expectedAllowed       func() gomegatypes.GomegaMatcher
		expectedFailureReason string
		expectedMessage       string
		expectedCode          int32
	}{
		// negative
		{
			name:               "deny delete operation for the etcd resource when it is currently in deletion",
			operation:          admissionv1.Delete,
			lastOperationType:  druidv1alpha1.LastOperationTypeDelete,
			lastOperationState: druidv1alpha1.LastOperationStateProcessing,
			expectedAllowed:    BeFalse,
			// either have just "operation DELETE is denied" or make this generic so the error string needn't be hardcoded
			expectedFailureReason: "Forbidden",
			expectedMessage:       "operation DELETE is denied: etcd deletion in progress",
			expectedCode:          http.StatusForbidden,
		},
		{
			name:                  "deny delete operation for the etcd resource when it is currently in reconciliation",
			operation:             admissionv1.Delete,
			lastOperationType:     druidv1alpha1.LastOperationTypeReconcile,
			lastOperationState:    druidv1alpha1.LastOperationStateProcessing,
			expectedAllowed:       BeFalse,
			expectedFailureReason: "Forbidden",
			// either have just "operation DELETE is denied" or make this generic so the error string needn't be hardcoded
			expectedMessage: "operation DELETE is denied: etcd reconciliation in progress",
			expectedCode:    http.StatusForbidden,
		},
		// positive
		{
			name:                  "allow delete operation for the etcd resource when it is not currently in reconciliation or deletion",
			operation:             admissionv1.Delete,
			lastOperationType:     druidv1alpha1.LastOperationTypeReconcile,
			lastOperationState:    druidv1alpha1.LastOperationStateSucceeded,
			expectedAllowed:       BeTrue,
			expectedFailureReason: "",
			expectedMessage:       "operation DELETE is allowed",
			expectedCode:          http.StatusOK,
		},
	}
	cl := testutils.CreateDefaultFakeClient()
	decoder := admission.NewDecoder(cl.Scheme())

	handler := &Handler{
		Client: cl,
		config: &Config{
			Enabled: true,
		},
		decoder: decoder,
		logger:  logr.Discard(),
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etcd := &druidv1alpha1.Etcd{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Etcd",
					APIVersion: "v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "etcd-test",
					Namespace: "etcd-test-namespace",
				},
				Status: druidv1alpha1.EtcdStatus{
					LastOperation: &druidv1alpha1.LastOperation{
						Type:  tc.lastOperationType,
						State: tc.lastOperationState,
					},
				},
			}
			bytes, err := json.Marshal(etcd)
			g.Expect(err).ToNot(HaveOccurred())

			rawExt := buildObjRawExtension(g, etcd, bytes, "", "", nil)

			resp := handler.Handle(context.Background(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: tc.operation,
					Kind:      etcdGVK,
					OldObject: rawExt,
				},
			})
			g.Expect(resp.Allowed).To(tc.expectedAllowed())
			g.Expect(resp.Result.Message).To(Equal(tc.expectedMessage))
			g.Expect(resp.Result.Code).To(Equal(tc.expectedCode))
		})
	}
}

// ---------------- Helper functions -------------------

func buildObjRawExtension(g *WithT, emptyObj runtime.Object, objRaw []byte, testObjectName, testNs string, labels map[string]string) runtime.RawExtension {
	var (
		rawBytes []byte
		err      error
	)
	rawBytes = objRaw
	obj := buildObject(getObjectGVK(g, emptyObj), testObjectName, testNs, labels)
	if objRaw == nil {
		rawBytes, err = json.Marshal(obj)
		g.Expect(err).ToNot(HaveOccurred())
	}

	// fmt.Println("The object that wasbuilt was: ", obj)
	// fmt.Println("The raw bytes were: ", string(rawBytes))
	return runtime.RawExtension{
		Object: obj,
		Raw:    rawBytes,
	}
}

func getObjectGVK(g *WithT, obj runtime.Object) schema.GroupVersionKind {
	gvk, err := apiutil.GVKForObject(obj, kubernetes.Scheme)
	g.Expect(err).ToNot(HaveOccurred())
	return gvk
}

func buildObject(gvk schema.GroupVersionKind, name, namespace string, labels map[string]string) runtime.Object {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)
	obj.SetLabels(labels)
	return obj
}
