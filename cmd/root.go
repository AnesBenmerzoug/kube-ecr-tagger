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
package cmd

import (
	"fmt"
	"log"
	"os"

	registry "github.com/AnesBenmerzoug/kube-ecr-tagger/internal/aws"
	"github.com/AnesBenmerzoug/kube-ecr-tagger/internal/helpers"
	"github.com/AnesBenmerzoug/kube-ecr-tagger/internal/kubernetes"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kube-ecr-tagger TAG",
	Short: "Tags images from ECR used by Pods in cluster",
	Long:  `A command that adds a given tag to all images from ECR that are used by Pods in the kubernetes cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(1)
		}
		err := findAndTagImages(args[0])
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
}

func findAndTagImages(tag string) error {
	ecrClient, err := registry.NewClient()
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewClient("")
	if err != nil {
		return err
	}

	log.Print("Finding all Pod images")

	imageNames, err := k8sClient.ListPodImages()
	if err != nil {
		return err
	}

	log.Printf("Found '%v' images", len(imageNames))

	log.Printf("Parsing image names")

	var ecrImages []*ecr.Image

	for _, imageName := range imageNames {
		image, err := helpers.ParseImageName(imageName)
		if err != nil {
			log.Print(err)
			continue
		}
		ecrImages = append(ecrImages, image)
	}

	if len(ecrImages) == 0 {
		return fmt.Errorf("No ECR images were found")
	}

	log.Printf("Found '%v' images from ECR", len(ecrImages))

	log.Printf("Getting information about found images from ECR")

	ecrImages, err = ecrClient.GetImagesInformation(ecrImages)
	if err != nil {
		return err
	}

	log.Printf("Tagging found ECR images with tag '%s'", tag)

	err = ecrClient.TagImages(ecrImages, tag)
	if err != nil {
		return err
	}

	return nil
}
