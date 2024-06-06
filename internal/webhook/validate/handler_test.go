// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	testutils "github.com/gardener/etcd-druid/test/utils"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

const (
	name      = "etcd-test"
	namespace = "shoot--foo--bar"
	uuid      = "f1a38edd-e506-412a-82e6-e0fa839d0707"
	provider  = "aws"
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
	g := NewWithT(t)

	testCases := []struct {
		name        string
		operation   admissionv1.Operation
		request     admissionv1.AdmissionRequest
		expectedMsg string
	}{
		// negative
		{
			name:        "deny delete operation for the etcd resource when it is currently in deletion",
			operation:   admissionv1.Delete,
			request:     admissionv1.AdmissionRequest{Operation: admissionv1.Delete},
			expectedMsg: "operation DELETE is denied",
		},
		{
			name:        "deny delete operation for the etcd resource when it is currently in reconciliation",
			operation:   admissionv1.Delete,
			expectedMsg: "operation DELETE is denied",
		},
		// positive
		{
			name:        "allow delete operation for the etcd resource when it is not currently in reconciliation or deletion",
			operation:   admissionv1.Delete,
			expectedMsg: "operation DELETE is allowed",
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

}

// func TestHandleCreate(t *testing.T) {
// g := NewWithT(t)

// testCases := []struct {
// 	name        string
// 	operation   admissionv1.Operation
// 	request     admission.Request
// 	expectedMsg string
// }{
// 	// negative tests
// 	{
// 		name:        "allow connect operation for the etcd resource",
// 		operation:   admissionv1.Create,
// 		request:     admission.Request{},
// 		expectedMsg: "operation CREATE is not allowed",
// 	},
// 	// positive tests
// 	{
// 		name:        "allow connect operation for the etcd resource",
// 		operation:   admissionv1.Create,
// 		request:     admission.Request{},
// 		expectedMsg: "operation CREATE is allowed",
// 	},
// }
// cl := testutils.CreateDefaultFakeClient()
// decoder := admission.NewDecoder(cl.Scheme())

// handler := &Handler{
// 	Client: cl,
// 	config: &Config{
// 		Enabled: true,
// 	},
// 	decoder: decoder,
// 	logger:  logr.Discard(),
// }
// }
