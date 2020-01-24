package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	registry "github.com/AnesBenmerzoug/kube-ecr-tagger/internal/ecr"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type mockECRClient struct {
	ecriface.ECRAPI
}

func (m *mockECRClient) BatchGetImage(input *ecr.BatchGetImageInput) (*ecr.BatchGetImageOutput, error) {
	var output ecr.BatchGetImageOutput
	for _, imageID := range input.ImageIds {
		output.Images = append(output.Images, &ecr.Image{
			ImageId:        imageID,
			RepositoryName: input.RepositoryName,
			RegistryId:     input.RegistryId,
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
		})
	}
	return &output, nil
}

func (m *mockECRClient) PutImage(input *ecr.PutImageInput) (*ecr.PutImageOutput, error) {
	output := ecr.PutImageOutput{
		Image: &ecr.Image{
			ImageId: &ecr.ImageIdentifier{
				ImageTag: input.ImageTag,
			},
			ImageManifest:  input.ImageManifest,
			RepositoryName: input.RepositoryName,
			RegistryId:     input.RegistryId,
		},
	}

	return &output, nil
}

func definePod(namespace, name, image string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: image},
			},
		},
	}
}

func definePodWithInitContainer(namespace, name, containerImage, initContainerImage string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{Image: initContainerImage},
			},
			Containers: []corev1.Container{
				{Image: containerImage},
			},
		},
	}
}

func TestTaggingPodImages(t *testing.T) {
	var tests = []struct {
		description string
		namespace   string
		name        string
		tag         string
		image       string
	}{
		{"non ecr image", "default", "pod", "test-tag", "test-image:latest"},
		{"ecr image different tag", "default", "pod", "test-tag", "123456789012.dkr.ecr.eu-central-1.amazonaws.com/test-image:latest"},
		{"ecr image already tagged", "default", "pod", "test-tag", "123456789012.dkr.ecr.us-west-2.amazonaws.com/test-image:test-tag"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Use a timeout to keep the test from hanging
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			// Create the fake kubernetes client
			client := fake.NewSimpleClientset()

			// Create fake ecr client
			// taken from:
			// https://github.com/aws/aws-sdk-go/blob/master/awstesting/mock/mock.go#L16,L26
			mockSession := func() *session.Session {
				// server is the mock server that simply writes a 200 status back to the client
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				return session.Must(session.NewSession(&aws.Config{
					DisableSSL: aws.Bool(true),
					Endpoint:   aws.String(server.URL),
					Region:     aws.String("eu-central-1"),
				}))
			}()
			ecrClient := &registry.Client{
				ECRAPI: &mockECRClient{
					ecr.New(mockSession),
				},
			}
			go func(ctx context.Context) {
				err := findAndTagImages(ctx, client, ecrClient, test.tag, test.namespace)
				if err != nil {
					t.Error(err)
				}
			}(ctx)

			time.Sleep(2 * time.Second)

			pod := definePod(test.namespace, test.name, test.image)
			_, err := client.CoreV1().Pods(test.namespace).Create(pod)
			if err != nil {
				t.Errorf("error creating pod: %v", err)
			}
			// add label to pod
			pod.ObjectMeta.Labels = make(map[string]string)
			pod.ObjectMeta.Labels["test"] = "test"
			_, err = client.CoreV1().Pods(test.namespace).Update(pod)
			if err != nil {
				t.Errorf("error updating pod: %v", err)
			}

			pod = definePodWithInitContainer(test.namespace, test.name+"-init", test.image, test.image)
			_, err = client.CoreV1().Pods(test.namespace).Create(pod)
			if err != nil {
				t.Errorf("error creating pod with init container: %v", err)
			}

			// add label to pod
			pod.ObjectMeta.Labels = make(map[string]string)
			pod.ObjectMeta.Labels["test"] = "test"
			_, err = client.CoreV1().Pods(test.namespace).Update(pod)
			if err != nil {
				t.Errorf("error updating pod with init container: %v", err)
			}

			time.Sleep(5 * time.Second)
		})
	}
}
