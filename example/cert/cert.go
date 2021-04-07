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
package main

import (
	"fmt"
	v1 "github.com/cuisongliu/kube-webhook/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func main() {
	certDir := os.TempDir() + "/webhook/serving-certs"
	obj := make(map[string]*metav1.LabelSelector)
	namespace := make(map[string]*metav1.LabelSelector)
	fmt.Printf("certDir: %s\n", certDir)
	w := &v1.CertWebHook{
		Subject:     nil, //证书数据
		CertDir:     certDir, //生成的证书位置
		Namespace:   "default", //秘钥以及webhook的svc的namespace
		ServiceName: "svcName", //生成webhook的对应service名称
		SecretName:  "certs", //存放证书名称
		CsrName:     "csr", //csr证书资源名称
		WebHook: []v1.WebHook{
			{MutatingName: "mutating-cfg", ObjectSelect: obj, NamespaceSelect: namespace},
			{ValidatingName: "validating-cfg", ObjectSelect: obj, NamespaceSelect: namespace},
		},
	}
	err := w.Init()
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		os.Exit(1)
	}
	err = w.Generator()
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		os.Exit(1)
	}
}
