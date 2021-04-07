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
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Defaulter defines functions for setting defaults on resources
type RuntimeObject interface {
	OutRuntimeObject() runtime.Object
	IntoRuntimeObject(runtime.Object)
	GetClient() client.Client
}

func JsonConvert(from interface{}, to interface{}) error {
	var data []byte
	var err error
	if data, err = json.Marshal(from); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(json.Unmarshal(data, to))
}

type WebhookObject struct {
	WK             *webhook.Server
	Webhook        RuntimeObject
	Obj            runtime.Object
	ValidatingPath string
	DefaultingPath string
	Client         client.Client
}

func (wko *WebhookObject) Init() {
	_ = wko.Webhook.(inject.Client).InjectClient(wko.Client)
	wko.Webhook.IntoRuntimeObject(wko.Obj)
	if v,ok:=wko.Webhook.(Validator);ok{
		wko.WK.Register(wko.ValidatingPath, ValidatingWebhookFor(v))
	}
	if m,ok:=wko.Webhook.(Defaulter);ok{
		wko.WK.Register(wko.DefaultingPath, DefaultingWebhookFor(m))
	}
}
