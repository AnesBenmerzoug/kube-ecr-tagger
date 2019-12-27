package helpers

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
)

var ecrRegex = regexp.MustCompile(`^(?P<registry>\d+)\.dkr\.ecr.\w+-\w+-\d\.amazonaws\.com/(?P<repository>.+):(?P<tag>.+)$`)

// ParseImageName parses a given ECR image name and extracts the registry ID, repository name and tag from it
func ParseImageName(imageName string) (*ecr.Image, error) {
	match := ecrRegex.FindStringSubmatch(imageName)
	if match != nil {
		return nil, fmt.Errorf("Could not parse image name '%s'", imageName)
	}
	image := &ecr.Image{
		ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String(match[2])},
		RepositoryName: aws.String(match[1]),
		RegistryId:     aws.String(match[0]),
	}
	return image, nil
}
