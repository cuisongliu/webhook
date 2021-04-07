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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func SetupWebhook(wk *webhook.Server, mgr ctrl.Manager) {
	c := mgr.GetClient()
	wkhpa := &v1.WebhookObject{
		WK:             wk,
		Webhook:        &HPAWebhook{},
		Obj:            &hpav1.HorizontalPodAutoscaler{},
		ValidatingPath: "/validate-autoscaling-v2beta1-hpa",
		DefaultingPath: "/mutate-autoscaling-v2beta1-hpa",
		Client:         c,
	}
	wkhpa.Init()
}
