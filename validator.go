/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Validator defines functions for validating an operation
type Validator interface {
	RuntimeObject
	ValidateCreate() error
	ValidateUpdate(old runtime.Object) error
	ValidateDelete() error
}

// ValidatingWebhookFor creates a new Webhook for validating the provided type.
func ValidatingWebhookFor(validator Validator) *admission.Webhook {
	return &admission.Webhook{
		Handler: &validatingHandler{validator: validator},
	}
}

type validatingHandler struct {
	validator Validator
	decoder   *admission.Decoder
}

var _ admission.DecoderInjector = &validatingHandler{}

// InjectDecoder injects the decoder into a validatingHandler.
func (h *validatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if h.validator == nil {
		panic("validator should never be nil")
	}
	// Get the object in the request
	if req.Operation == v1beta1.Create {
		into := &unstructured.Unstructured{}
		err := h.decoder.Decode(req, into)
		h.validator.IntoRuntimeObject(into)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.validator.ValidateCreate()
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Update {
		oldObj := h.validator.OutRuntimeObject().DeepCopyObject()
		into := &unstructured.Unstructured{}
		err := h.decoder.DecodeRaw(req.Object, into)
		h.validator.IntoRuntimeObject(into)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateUpdate(oldObj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1beta1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		into := &unstructured.Unstructured{}
		err := h.decoder.DecodeRaw(req.OldObject, into)
		h.validator.IntoRuntimeObject(into)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = h.validator.ValidateDelete()
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}
