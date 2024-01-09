//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package stern

import (
	"context"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// List returns a map of all 'app.kubernetes.io/instance' values.
func List(ctx context.Context, client kubernetes.Interface, config *Config) (map[string]string, error) {
	var namespaces []string
	// A specific namespace is ignored if all-namespaces is provided.
	if config.AllNamespaces {
		namespaces = []string{""}
	} else {
		namespaces = config.Namespaces
	}

	labels := make(map[string]string)
	options := metav1.ListOptions{}

	wg := sync.WaitGroup{}

	wg.Add(len(namespaces))

	// Concurrently iterate through provided namespaces.
	for _, n := range namespaces {
		go func(n string) {
			defer wg.Done()

			pods, err := client.CoreV1().Pods(n).List(ctx, options)
			if err != nil {
				return
			}

			match := "app.kubernetes.io/instance"
			// Iterate through pods in namespace, looking for matching labels.
			for _, pod := range pods.Items {
				key := pod.Labels[match]

				if key == "" {
					continue
				}

				labels[key] = match
			}
		}(n)
	}

	wg.Wait()

	return labels, nil
}

// ListTargets returns targets by listing and filtering pods
func ListTargets(ctx context.Context, i corev1client.PodInterface, labelSelector labels.Selector, fieldSelector fields.Selector, filter *targetFilter) ([]*Target, error) {
	list, err := i.List(ctx, metav1.ListOptions{LabelSelector: labelSelector.String(), FieldSelector: fieldSelector.String()})
	if err != nil {
		return nil, err
	}
	var targets []*Target
	for i := range list.Items {
		filter.visit(&list.Items[i], func(t *Target) {
			targets = append(targets, t)
		})
	}
	return targets, nil
}
