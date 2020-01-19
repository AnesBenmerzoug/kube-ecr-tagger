/*
Copyright © 2019 Anes Benmerzoug

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
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"

	registry "github.com/AnesBenmerzoug/kube-ecr-tagger/internal/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type commandOpts struct {
	Namespace string
}

var opts = commandOpts{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kube-ecr-tagger TAG",
	Short: "Tags images from ECR used by Pods in cluster",
	Long:  `A command that adds a given tag to all images from ECR that are used by Pods in the kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(1)
		}
		err := findAndTagImages(args[0], opts)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	rootCmd.Flags().StringVar(&opts.Namespace, "namespace", corev1.NamespaceAll, "namespace from which images will be listed. Defaults to all namespaces")
}

func findAndTagImages(tag string, opts commandOpts) error {
	ecrClient, err := registry.NewClient()
	if err != nil {
		return err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	defaultResyncPeriod := 0 * time.Second
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, defaultResyncPeriod, informers.WithNamespace(opts.Namespace))
	informer := factory.Core().V1().Pods().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			tagPodImages(ecrClient, tag, obj)
		},
		UpdateFunc: func(new interface{}, old interface{}) {
			tagPodImages(ecrClient, tag, new)
		},
	})
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		err := fmt.Errorf("Timed out waiting for caches to sync")
		runtime.HandleError(err)
		return err
	}
	<-stopper

	return nil
}

func tagPodImages(ecrClient *registry.Client, tag string, obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	var ecrImages []*ecr.Image
	// Get from init containers all images that are from ECR
	for _, container := range pod.Spec.InitContainers {
		image, err := registry.ParseImageName(container.Image)
		if err != nil {
			log.Print(err)
			continue
		}
		ecrImages = append(ecrImages, image)
	}
	// Get from containers all images that are from ECR
	for _, container := range pod.Spec.Containers {
		image, err := registry.ParseImageName(container.Image)
		if err != nil {
			log.Print(err)
			continue
		}
		ecrImages = append(ecrImages, image)
	}
	// Get the images' manifests from ECR
	ecrImages, err := ecrClient.GetImagesInformation(ecrImages)
	if err != nil {
		return
	}
	// Add the given tag to all images
	err = ecrClient.TagImages(ecrImages, tag)
	if err != nil {
		return
	}
}
