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
	v1 "github.com/cuisongliu/webhook"
	hpav1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var hpalog = logf.Log.WithName("hpa-resource")

type HPAWebhook struct {
	object *hpav1.HorizontalPodAutoscaler
	client client.Client
}

func (a *HPAWebhook) OutRuntimeObject() runtime.Object {
	return a.object
}
func (a *HPAWebhook) GetClient() client.Client {
	return a.client
}
func (r *HPAWebhook) IntoRuntimeObject(object runtime.Object) {
	obj := &hpav1.HorizontalPodAutoscaler{}
	_ = v1.JsonConvert(object, obj)
	r.object = obj
}

func (a *HPAWebhook) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-autoscaling-v2beta1-hpa,mutating=true,failurePolicy=fail,groups=autoscaling,resources=horizontalpodautoscalers,verbs=create;update,versions=v2beta1,name=mhorizontalpodautoscaler.kb.io

var _ v1.Defaulter = &HPAWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HPAWebhook) Default(req admission.Request) {
	hpalog.Info("default", "name", r.object.Name)
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-autoscaling-v2beta1-hpa,mutating=false,failurePolicy=fail,groups=autoscaling,resources=horizontalpodautoscalers,versions=v2beta1,name=vhorizontalpodautoscaler.kb.io

var _ v1.Validator = &HPAWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HPAWebhook) ValidateCreate(req admission.Request) error {
	hpalog.Info("validate create", "name", r.object.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HPAWebhook) ValidateUpdate(req admission.Request) error {
	hpalog.Info("validate update", "name", r.object.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HPAWebhook) ValidateDelete(req admission.Request) error {
	hpalog.Info("validate delete", "name", r.object.Name)
	return nil
}
