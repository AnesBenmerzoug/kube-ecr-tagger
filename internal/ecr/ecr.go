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
package registry

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

var ecrRegex = regexp.MustCompile(`^(?P<registry>\d+)\.dkr\.ecr.\w+-\w+-\d\.amazonaws\.com/(?P<repository>.+):(?P<tag>.+)$`)

// Client wraps an ECR API
type Client struct {
	ecriface.ECRAPI
}

// NewClient instantiates a new Client struct
func NewClient() (*Client, error) {
	config := aws.NewConfig()

	currentSession, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		ecr.New(currentSession),
	}

	return client, nil
}

// GetImageTags queries ECR to get all Tags for the given image
func (c *Client) GetImageTags(image *ecr.Image) ([]*string, error) {
	var imageTags []*string
	describeInput := &ecr.DescribeImagesInput{
		ImageIds: []*ecr.ImageIdentifier{
			{
				ImageTag: image.ImageId.ImageTag,
			},
		},
		RepositoryName: image.RepositoryName,
		RegistryId:     image.RegistryId,
	}
	result, err := c.DescribeImages(describeInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Print(aerr.Error())
			return nil, err
		} else {
			return nil, err
		}
	}
	for _, imageDetail := range result.ImageDetails {
		imageTags = append(imageTags, imageDetail.ImageTags...)
	}
	return imageTags, nil
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
				log.Print(aerr.Error())
				continue
			} else {
				return nil, err
			}
		}
		imageInformation = append(imageInformation, result.Images...)
	}
	return imageInformation, nil
}

// TagImages adds the given tag to a list of images on ECR
func (c *Client) TagImages(imagesToTag []*ecr.Image, tag string) error {
	for _, image := range imagesToTag {
		if *image.ImageId.ImageTag == tag {
			log.Printf("Image '%s' already has tag '%s'", image.ImageId.String(), *image.ImageId.ImageTag)
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
				log.Print(aerr.Error())
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

// ParseImageName parses a given ECR image name and extracts the registry ID, repository name and tag from it
func ParseImageName(imageName string) (*ecr.Image, error) {
	match := ecrRegex.FindStringSubmatch(imageName)
	if match == nil {
		return nil, fmt.Errorf("Could not parse image name '%s'", imageName)
	}
	image := &ecr.Image{
		ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String(match[2])},
		RepositoryName: aws.String(match[1]),
		RegistryId:     aws.String(match[0]),
	}
	return image, nil
}
