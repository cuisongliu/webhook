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
package v1

import (
	"flag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

type WebHook struct {
	ValidatingName  string
	MutatingName    string
	ObjectSelect    map[string]*v1.LabelSelector
	NamespaceSelect map[string]*v1.LabelSelector
}

type CertWebHook struct {
	//证书相关
	Subject []string
	CertDir string
	//kubernetes相关资源
	Namespace   string
	ServiceName string
	SecretName  string
	CsrName     string
	WebHook     []WebHook

	client *kubernetes.Clientset
}

func newK8sClient() (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	config, err = rest.InClusterConfig()
	if err != nil {
		var kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	// creates the clientSet
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *CertWebHook) Init() error {
	if c.Subject == nil || len(c.Subject) == 0 {
		c.Subject = []string{"cuisongliu CN"}
	}
	if c.Namespace == "" {
		c.Namespace = "kube-system"
	}
	if c.CertDir == "" {
		c.CertDir = os.TempDir() + "/k8s-webhook-server/serving-certs"
	}
	if c.ServiceName == "" {
		c.ServiceName = "webhook-service"
	}
	if c.SecretName == "" {
		c.SecretName = "webhook-secret"
	}
	if c.CsrName == "" {
		c.CsrName = "webhook-csr"
	}
	if c.WebHook == nil || len(c.WebHook) == 0 {
		c.WebHook = []WebHook{
			{MutatingName: "mutating-webhook-configuration"},
			{ValidatingName: "validating-webhook-configuration"},
		}
	}
	var err error

	c.client, err = newK8sClient()
	if err != nil {
		return err
	}
	return nil

}

func (c *CertWebHook) Generator() error {
	secret, err := c.generateSecret()
	if err != nil {
		return err
	}
	if err := c.patchWebHook(string(secret.Data[caBundleKey])); err != nil {
		return err
	}
	if err := c.writeTLSFiles(secret.Data[certKey], secret.Data[keyKey]); err != nil {
		return err
	}
	return nil
}
