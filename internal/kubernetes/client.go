/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps a kubernetes clientset
type Client struct {
	clientset kubernetes.Interface
}

// NewClient instantiates a new Client struct from either a kubeconfig file (out of cluster)
// or a service account token (in-cluster)
func NewClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		clientset: clientset,
	}

	return client, nil
}

// ListImages returns a list of images currently used by Pods
func (c *Client) ListImages(namespace string) ([]string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	imageMap := make(map[string]struct{}, 8)
	var images []string

	for _, pod := range pods.Items {
		for _, initContainer := range pod.Spec.InitContainers {
			if _, exists := imageMap[initContainer.Image]; !exists {
				imageMap[initContainer.Image] = struct{}{}
				images = append(images, initContainer.Image)
			}
		}
		for _, container := range pod.Spec.Containers {
			if _, exists := imageMap[container.Image]; !exists {
				imageMap[container.Image] = struct{}{}
				images = append(images, container.Image)
			}
		}
	}

	return images, nil
}
