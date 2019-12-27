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
package registry

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

// Client wraps an ECR API
type Client struct {
	ecriface.ECRAPI
}

// NewClient instantiates a new Client struct
func NewClient() (*Client, error) {
	config := aws.NewConfig()

	currentSession := session.New(config)

	awsCredentials := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(currentSession),
			},
		})

	config.WithCredentials(awsCredentials)

	client := &Client{
		ecr.New(currentSession),
	}

	return client, nil
}

// GetImagesInformation queries ECR to get information for the given images
func (c *Client) GetImagesInformation(images []*ecr.Image) ([]*ecr.Image, error) {
	var imageInformation []*ecr.Image
	for _, image := range images {
		getInput := &ecr.BatchGetImageInput{
			ImageIds: []*ecr.ImageIdentifier{
				{
					ImageTag: image.ImageId.ImageTag,
				},
			},
			RepositoryName: image.RepositoryName,
			RegistryId:     image.RegistryId,
		}
		result, err := c.BatchGetImage(getInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Printf(aerr.Error())
			} else {
				return nil, err
			}
			continue
		}
		for _, resultImage := range result.Images {
			imageInformation = append(imageInformation, resultImage)
		}
	}
	return imageInformation, nil
}

// TagImages adds the given tag to a list of images on ECR
func (c *Client) TagImages(imagesToTag []*ecr.Image, tag string) error {
	for _, image := range imagesToTag {
		if *image.ImageId.ImageTag == tag {
			continue
		}
		putInput := &ecr.PutImageInput{
			ImageManifest:  image.ImageManifest,
			ImageTag:       aws.String(tag),
			RepositoryName: image.RepositoryName,
			RegistryId:     image.RegistryId,
		}
		_, err := c.PutImage(putInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Printf(aerr.Error())
			} else {
				return err
			}
		}
	}
	return nil
}
