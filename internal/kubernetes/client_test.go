package k8s

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func pod(namespace, image string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Image: image},
			},
		},
	}
}

func podWithInitContainer(namespace, containerImage, initContainerImage string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{Image: initContainerImage},
			},
			Containers: []v1.Container{
				{Image: containerImage},
			},
		},
	}
}

func TestListImages(t *testing.T) {
	var tests = []struct {
		description string
		namespace   string
		expected    []string
		objs        []runtime.Object
	}{
		{"no pods", "", nil, nil},
		{"all namespaces", "", []string{"a", "b"}, []runtime.Object{pod("correct-namespace", "a"), pod("wrong-namespace", "b")}},
		{"filter namespace", "correct-namespace", []string{"a"}, []runtime.Object{pod("correct-namespace", "a"), pod("wrong-namespace", "b")}},
		{"wrong namespace", "correct-namespace", nil, []runtime.Object{pod("wrong-namespace", "b")}},
		{"pod with initContainer same image", "", []string{"c"}, []runtime.Object{podWithInitContainer("default", "c", "c")}},
		{"pod with initContainer different image", "", []string{"d", "c"}, []runtime.Object{podWithInitContainer("default", "c", "d")}},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				clientset: fake.NewSimpleClientset(test.objs...),
			}

			actual, err := client.ListImages(test.namespace)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(actual, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}
