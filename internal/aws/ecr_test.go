package registry

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/mock"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/google/go-cmp/cmp"
)

type mockBatchGetImageClient struct {
	ecriface.ECRAPI
	response ecr.BatchGetImageOutput
}

func (m *mockBatchGetImageClient) BatchGetImage(input *ecr.BatchGetImageInput) (*ecr.BatchGetImageOutput, error) {
	return &m.response, nil
}

func TestGettingImageInformation(t *testing.T) {
	var tests = []struct {
		description string
		input       []*ecr.Image
		response    ecr.BatchGetImageOutput
		expected    []*ecr.Image
	}{
		{"no image - 1", nil, ecr.BatchGetImageOutput{}, nil},
		{"no image - 2", []*ecr.Image{}, ecr.BatchGetImageOutput{}, nil},
		{"image",
			[]*ecr.Image{
				{
					ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String("latest")},
					RepositoryName: aws.String("test"),
					RegistryId:     aws.String("530519006690"),
				},
			},
			ecr.BatchGetImageOutput{
				Images: []*ecr.Image{
					{
						ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String("latest")},
						RepositoryName: aws.String("test"),
						RegistryId:     aws.String("530519006690"),
						ImageManifest: aws.String(`{
"schemaVersion": 2,
"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
"config": {
	"mediaType": "application/vnd.docker.container.image.v1+json",
	"size": 7023,
	"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
},
"layers": [
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 32654,
		"digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	},
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 16724,
		"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
	},
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 73109,
		"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
	}
]
}
`),
					},
				},
			}, []*ecr.Image{
				{
					ImageId:        &ecr.ImageIdentifier{ImageTag: aws.String("latest")},
					RepositoryName: aws.String("test"),
					RegistryId:     aws.String("530519006690"),
					ImageManifest: aws.String(`{
"schemaVersion": 2,
"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
"config": {
	"mediaType": "application/vnd.docker.container.image.v1+json",
	"size": 7023,
	"digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
},
"layers": [
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 32654,
		"digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
	},
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 16724,
		"digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
	},
	{
		"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
		"size": 73109,
		"digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
	}
]
}
`),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockSession := mock.Session
			client := &Client{
				&mockBatchGetImageClient{
					ecr.New(mockSession),
					test.response,
				},
			}
			actual, err := client.GetImagesInformation(test.input)
			if test.expected == nil && actual != nil && err == nil {
				t.Errorf("Expected no image, but got '%+v' instead", actual)
				return
			}
			if diff := cmp.Diff(actual, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}

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
