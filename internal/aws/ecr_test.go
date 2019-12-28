package registry

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func TestParseImageName(t *testing.T) {
	var tests = []struct {
		description string
		imageName   string
		expected    *ecr.Image
	}{
		{"no image", "", nil},
		{"non ecr image", "golang:latest", nil},
		{"ecr image",
			"530519006690.dkr.ecr.eu-central-1.amazonaws.com/test:latest",
			&ecr.Image{RegistryId: aws.String("530519006690"),
				RepositoryName: aws.String("test"),
				ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String("latest")}},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			image, err := ParseImageName(test.imageName)
			if test.expected == nil && image != nil && err == nil {
				t.Errorf("Expected no image, but got '%+v' instead", image)
				return
			}

		})
	}
}
