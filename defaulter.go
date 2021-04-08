/*
Copyright © 2021 cuisongliu@qq.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package webhook

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Defaulter defines functions for setting defaults on resources
type Defaulter interface {
	Default(req admission.Request)
	RuntimeObject
}

// DefaultingWebhookFor creates a new Webhook for Defaulting the provided type.
func DefaultingWebhookFor(defaulter Defaulter) *admission.Webhook {
	return &admission.Webhook{
		Handler: &mutatingHandler{defaulter: defaulter},
	}
}

type mutatingHandler struct {
	defaulter Defaulter
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &mutatingHandler{}

// InjectDecoder injects the decoder into a mutatingHandler.
func (h *mutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.defaulter == nil {
		panic("callback should never be nil")
	}

	// Get the object in the request
	//obj := h.callback(h.defaulter.OutRuntimeObject().DeepCopyObject(), h.defaulter.GetClient())
	into := &unstructured.Unstructured{}
	err := h.decoder.Decode(req, into)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.defaulter.IntoRuntimeObject(into)
	// Default the object
	h.defaulter.Default(req)
	marshalled, err := json.Marshal(h.defaulter.OutRuntimeObject())
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}
