package printer_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/apps-cli-plugin/pkg/apis"
	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/printer"
)

func TestExportResource(t *testing.T) {
	scheme := runtime.NewScheme()
	cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name        string
		obj         printer.Object
		format      printer.OutputFormat
		want        string
		shouldError bool
	}{{

		name:   "empty",
		format: printer.OutputFormatYaml,
		obj:    &cartov1alpha1.Workload{},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata: {}
spec: {}
`,
	}, {
		name:   "export named",
		format: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-workload",
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  name: my-workload
spec: {}
`,
	}, {
		name:   "export named in json",
		format: printer.OutputFormatJson,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-workload",
			},
		},
		want: `
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"name": "my-workload"
	},
	"spec": {}
}
`,
	}, {
		name:   "export generated named",
		format: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:         "my-workload-abcde",
				GenerateName: "my-workload-",
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  generateName: my-workload-
spec: {}
`,
	}, {
		name:   "prune metadata",
		format: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				// fields to keep
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				// fields to drop
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Now(),
				DeletionTimestamp:          &metav1.Time{Time: time.Now()},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				ClusterName: "my-cluster",
				SelfLink:    "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  annotations:
    name: value
  labels:
    name: value
  name: my-workload
  namespace: default
spec: {}
`,
	}, {
		name:   "prune metadata export in json",
		format: printer.OutputFormatJson,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				// fields to keep
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				// fields to drop
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Now(),
				DeletionTimestamp:          &metav1.Time{Time: time.Now()},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				ClusterName: "my-cluster",
				SelfLink:    "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
		},
		want: `
{
	"apiVersion": "carto.run/v1alpha1",
	"kind": "Workload",
	"metadata": {
		"annotations": {
			"name": "value"
		},
		"labels": {
			"name": "value"
		},
		"name": "my-workload",
		"namespace": "default"
	},
	"spec": {}
}
`,
	}, {
		name:   "keep spec",
		format: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			Spec: cartov1alpha1.WorkloadSpec{
				Image: "my-image",
				Env: []corev1.EnvVar{
					{
						Name:  "NAME",
						Value: "value",
					},
				},
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata: {}
spec:
  env:
  - name: NAME
    value: value
  image: my-image
`,
	}, {
		name:   "drop status",
		format: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    cartov1alpha1.WorkloadConditionReady,
						Status:  metav1.ConditionTrue,
						Reason:  "No printing status",
						Message: "a hopefully informative message about what went wrong",
						LastTransitionTime: metav1.Time{
							Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
						},
					},
				},
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata: {}
spec: {}
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := printer.ExportResource(test.obj, test.format, scheme)
			if (err != nil) != test.shouldError {
				t.Errorf("ExportResource() error = %v, expected %v", err, test.shouldError)
			}
			if diff := cmp.Diff(strings.TrimSpace(test.want), got); diff != "" {
				t.Errorf("ExportResource() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestOutputResource(t *testing.T) {
	scheme := runtime.NewScheme()
	cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name         string
		obj          printer.Object
		want         string
		shouldError  bool
		outputFormat printer.OutputFormat
	}{{
		name:         "print output with yaml",
		outputFormat: printer.OutputFormatYaml,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				ClusterName: "my-cluster",
				SelfLink:    "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    cartov1alpha1.WorkloadConditionReady,
						Status:  metav1.ConditionTrue,
						Reason:  "No printing status",
						Message: "a hopefully informative message about what went wrong",
						LastTransitionTime: metav1.Time{
							Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
						},
					},
				},
			},
		},
		want: `
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  annotations:
    name: value
  clusterName: my-cluster
  creationTimestamp: "2021-09-10T15:00:00Z"
  deletionGracePeriodSeconds: 5
  deletionTimestamp: "2021-09-10T15:00:00Z"
  finalizers:
  - my.finalizer
  generation: 1
  labels:
    name: value
  managedFields:
  - manager: tanzu
  name: my-workload
  namespace: default
  ownerReferences:
  - apiVersion: v1
    kind: Pod
    name: workload-owner
    uid: ""
  resourceVersion: "999"
  selfLink: /default/my-workload
  uid: uid-xyz
spec: {}
status:
  conditions:
  - lastTransitionTime: "2019-06-29T01:44:05Z"
    message: a hopefully informative message about what went wrong
    reason: No printing status
    status: "True"
    type: Ready
  supplyChainRef: {}
`,
	}, {
		name:         "print output with json",
		outputFormat: printer.OutputFormatJson,
		obj: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload",
				Namespace: "default",
				Annotations: map[string]string{
					"name": "value",
				},
				Labels: map[string]string{
					"name": "value",
				},
				UID:                        types.UID("uid-xyz"),
				ResourceVersion:            "999",
				Generation:                 1,
				CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
				DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
				Finalizers:                 []string{"my.finalizer"},
				DeletionGracePeriodSeconds: &[]int64{5}[0],
				ManagedFields: []metav1.ManagedFieldsEntry{
					{Manager: "tanzu"},
				},
				ClusterName: "my-cluster",
				SelfLink:    "/default/my-workload",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "workload-owner",
					},
				},
			},
			Status: cartov1alpha1.WorkloadStatus{
				Conditions: []metav1.Condition{
					{
						Type:    cartov1alpha1.WorkloadConditionReady,
						Status:  metav1.ConditionTrue,
						Reason:  "No printing status",
						Message: "a hopefully informative message about what went wrong",
						LastTransitionTime: metav1.Time{
							Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
						},
					},
				},
			},
		},
		want: `
{
	"kind": "Workload",
	"apiVersion": "carto.run/v1alpha1",
	"metadata": {
		"name": "my-workload",
		"namespace": "default",
		"selfLink": "/default/my-workload",
		"uid": "uid-xyz",
		"resourceVersion": "999",
		"generation": 1,
		"creationTimestamp": "2021-09-10T15:00:00Z",
		"deletionTimestamp": "2021-09-10T15:00:00Z",
		"deletionGracePeriodSeconds": 5,
		"labels": {
			"name": "value"
		},
		"annotations": {
			"name": "value"
		},
		"ownerReferences": [
			{
				"apiVersion": "v1",
				"kind": "Pod",
				"name": "workload-owner",
				"uid": ""
			}
		],
		"finalizers": [
			"my.finalizer"
		],
		"clusterName": "my-cluster",
		"managedFields": [
			{
				"manager": "tanzu"
			}
		]
	},
	"spec": {},
	"status": {
		"conditions": [
			{
				"type": "Ready",
				"status": "True",
				"lastTransitionTime": "2019-06-29T01:44:05Z",
				"reason": "No printing status",
				"message": "a hopefully informative message about what went wrong"
			}
		],
		"supplyChainRef": {}
	}
}
`,
	}, {
		name:         "not valid output",
		outputFormat: "myFormat",
		obj:          &cartov1alpha1.Workload{},
		shouldError:  true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := printer.OutputResource(test.obj, test.outputFormat, scheme)
			if (err != nil) != test.shouldError {
				t.Errorf("OutputResource() error = %v, expected %v", err, test.shouldError)
			}
			if diff := cmp.Diff(strings.TrimSpace(test.want), got); diff != "" {
				t.Errorf("OutputResource() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestOutputResources(t *testing.T) {
	scheme := runtime.NewScheme()
	cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name         string
		objs         []printer.Object
		want         string
		shouldError  bool
		outputFormat printer.OutputFormat
	}{{
		name:         "print output with yaml",
		outputFormat: printer.OutputFormatYaml,
		objs: []printer.Object{
			&cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-workload",
					Namespace: "default",
					Annotations: map[string]string{
						"name": "value",
					},
					Labels: map[string]string{
						"name": "value",
					},
					UID:                        types.UID("uid-xyz"),
					ResourceVersion:            "999",
					Generation:                 1,
					CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
					Finalizers:                 []string{"my.finalizer"},
					DeletionGracePeriodSeconds: &[]int64{5}[0],
					ManagedFields: []metav1.ManagedFieldsEntry{
						{Manager: "tanzu"},
					},
					ClusterName: "my-cluster",
					SelfLink:    "/default/my-workload",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       "workload-owner",
						},
					},
				},
				Status: cartov1alpha1.WorkloadStatus{
					Conditions: []metav1.Condition{
						{
							Type:    cartov1alpha1.WorkloadConditionReady,
							Status:  metav1.ConditionTrue,
							Reason:  "No printing status",
							Message: "a hopefully informative message about what went wrong",
							LastTransitionTime: metav1.Time{
								Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
							},
						},
					},
				},
			},
			&cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "another-workload",
					Namespace: "default",
					Labels: map[string]string{
						"name": "value",
					},
					UID:                        types.UID("uid-xyz"),
					ResourceVersion:            "999",
					Generation:                 1,
					CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
					Finalizers:                 []string{"my.finalizer"},
					DeletionGracePeriodSeconds: &[]int64{5}[0],
					ManagedFields: []metav1.ManagedFieldsEntry{
						{Manager: "tanzu"},
					},
					ClusterName: "my-cluster",
					SelfLink:    "/default/my-workload",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       "workload-owner",
						},
					},
				},
				Status: cartov1alpha1.WorkloadStatus{
					Conditions: []metav1.Condition{
						{
							Type:    cartov1alpha1.WorkloadConditionReady,
							Status:  metav1.ConditionTrue,
							Reason:  "No printing status",
							Message: "a hopefully informative message about what went wrong",
							LastTransitionTime: metav1.Time{
								Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
							},
						},
					},
				},
			},
		},
		want: `
---
- apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    annotations:
      name: value
    clusterName: my-cluster
    creationTimestamp: "2021-09-10T15:00:00Z"
    deletionGracePeriodSeconds: 5
    deletionTimestamp: "2021-09-10T15:00:00Z"
    finalizers:
    - my.finalizer
    generation: 1
    labels:
      name: value
    managedFields:
    - manager: tanzu
    name: my-workload
    namespace: default
    ownerReferences:
    - apiVersion: v1
      kind: Pod
      name: workload-owner
      uid: ""
    resourceVersion: "999"
    selfLink: /default/my-workload
    uid: uid-xyz
  spec: {}
  status:
    conditions:
    - lastTransitionTime: "2019-06-29T01:44:05Z"
      message: a hopefully informative message about what went wrong
      reason: No printing status
      status: "True"
      type: Ready
    supplyChainRef: {}
- apiVersion: carto.run/v1alpha1
  kind: Workload
  metadata:
    clusterName: my-cluster
    creationTimestamp: "2021-09-10T15:00:00Z"
    deletionGracePeriodSeconds: 5
    deletionTimestamp: "2021-09-10T15:00:00Z"
    finalizers:
    - my.finalizer
    generation: 1
    labels:
      name: value
    managedFields:
    - manager: tanzu
    name: another-workload
    namespace: default
    ownerReferences:
    - apiVersion: v1
      kind: Pod
      name: workload-owner
      uid: ""
    resourceVersion: "999"
    selfLink: /default/my-workload
    uid: uid-xyz
  spec: {}
  status:
    conditions:
    - lastTransitionTime: "2019-06-29T01:44:05Z"
      message: a hopefully informative message about what went wrong
      reason: No printing status
      status: "True"
      type: Ready
    supplyChainRef: {}
`}, {
		name:         "print output with json",
		outputFormat: printer.OutputFormatJson,
		objs: []printer.Object{
			&cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-workload",
					Namespace: "default",
					Labels: map[string]string{
						"name": "value",
					},
					UID:                        types.UID("uid-xyz"),
					ResourceVersion:            "999",
					Generation:                 1,
					CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
					Finalizers:                 []string{"my.finalizer"},
					DeletionGracePeriodSeconds: &[]int64{5}[0],
					ManagedFields: []metav1.ManagedFieldsEntry{
						{Manager: "tanzu"},
					},
					ClusterName: "my-cluster",
					SelfLink:    "/default/my-workload",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       "workload-owner",
						},
					},
				},
				Status: cartov1alpha1.WorkloadStatus{
					Conditions: []metav1.Condition{
						{
							Type:    cartov1alpha1.WorkloadConditionReady,
							Status:  metav1.ConditionTrue,
							Reason:  "No printing status",
							Message: "a hopefully informative message about what went wrong",
							LastTransitionTime: metav1.Time{
								Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
							},
						},
					},
				},
			},
			&cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "another-workload",
					Namespace: "default",
					Annotations: map[string]string{
						"name": "value",
					},
					UID:                        types.UID("uid-abc"),
					ResourceVersion:            "1000",
					Generation:                 1,
					CreationTimestamp:          metav1.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC),
					DeletionTimestamp:          &metav1.Time{Time: time.Date(2021, time.September, 10, 15, 00, 00, 00, time.UTC)},
					Finalizers:                 []string{"my.finalizer"},
					DeletionGracePeriodSeconds: &[]int64{5}[0],
					ManagedFields: []metav1.ManagedFieldsEntry{
						{Manager: "tanzu"},
					},
					ClusterName: "my-cluster",
					SelfLink:    "/default/my-workload",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       "workload-owner",
						},
					},
				},
				Status: cartov1alpha1.WorkloadStatus{
					Conditions: []metav1.Condition{
						{
							Type:    cartov1alpha1.WorkloadConditionReady,
							Status:  metav1.ConditionTrue,
							Reason:  "No printing status",
							Message: "a hopefully informative message about what went wrong",
							LastTransitionTime: metav1.Time{
								Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
							},
						},
					},
				},
			},
		},
		want: `
[
	{
		"kind": "Workload",
		"apiVersion": "carto.run/v1alpha1",
		"metadata": {
			"name": "my-workload",
			"namespace": "default",
			"selfLink": "/default/my-workload",
			"uid": "uid-xyz",
			"resourceVersion": "999",
			"generation": 1,
			"creationTimestamp": "2021-09-10T15:00:00Z",
			"deletionTimestamp": "2021-09-10T15:00:00Z",
			"deletionGracePeriodSeconds": 5,
			"labels": {
				"name": "value"
			},
			"ownerReferences": [
				{
					"apiVersion": "v1",
					"kind": "Pod",
					"name": "workload-owner",
					"uid": ""
				}
			],
			"finalizers": [
				"my.finalizer"
			],
			"clusterName": "my-cluster",
			"managedFields": [
				{
					"manager": "tanzu"
				}
			]
		},
		"spec": {},
		"status": {
			"conditions": [
				{
					"type": "Ready",
					"status": "True",
					"lastTransitionTime": "2019-06-29T01:44:05Z",
					"reason": "No printing status",
					"message": "a hopefully informative message about what went wrong"
				}
			],
			"supplyChainRef": {}
		}
	},
	{
		"kind": "Workload",
		"apiVersion": "carto.run/v1alpha1",
		"metadata": {
			"name": "another-workload",
			"namespace": "default",
			"selfLink": "/default/my-workload",
			"uid": "uid-abc",
			"resourceVersion": "1000",
			"generation": 1,
			"creationTimestamp": "2021-09-10T15:00:00Z",
			"deletionTimestamp": "2021-09-10T15:00:00Z",
			"deletionGracePeriodSeconds": 5,
			"annotations": {
				"name": "value"
			},
			"ownerReferences": [
				{
					"apiVersion": "v1",
					"kind": "Pod",
					"name": "workload-owner",
					"uid": ""
				}
			],
			"finalizers": [
				"my.finalizer"
			],
			"clusterName": "my-cluster",
			"managedFields": [
				{
					"manager": "tanzu"
				}
			]
		},
		"spec": {},
		"status": {
			"conditions": [
				{
					"type": "Ready",
					"status": "True",
					"lastTransitionTime": "2019-06-29T01:44:05Z",
					"reason": "No printing status",
					"message": "a hopefully informative message about what went wrong"
				}
			],
			"supplyChainRef": {}
		}
	}
]`,
	}, {
		name:         "empty list with json format",
		outputFormat: printer.OutputFormatJson,
		objs:         []printer.Object{},
		want:         "[]",
	}, {
		name:         "empty list with yaml format",
		outputFormat: printer.OutputFormatYaml,
		objs:         []printer.Object{},
		want: `---
[]`,
	}, {
		name:         "not valid output",
		outputFormat: "myFormat",
		objs: []printer.Object{
			&cartov1alpha1.Workload{},
		},
		shouldError: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := printer.OutputResources(test.objs, test.outputFormat, scheme)
			if (err != nil) != test.shouldError {
				t.Errorf("OutputResources() error = %v, expected %v", err, test.shouldError)
			}
			if diff := cmp.Diff(strings.TrimSpace(test.want), got); diff != "" {
				t.Errorf("OutputResources() (-want, +got) = %v", diff)
			}
		})
	}
}

func TestResourceDiff(t *testing.T) {
	scheme := runtime.NewScheme()
	cartov1alpha1.AddToScheme(scheme)

	tests := []struct {
		name        string
		scheme      *runtime.Scheme
		left        printer.Object
		right       printer.Object
		want        string
		noChange    bool
		shouldError bool
	}{{
		name:   "output differences with +/-",
		scheme: scheme,
		left: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "change",
				Labels: map[string]string{
					apis.AppPartOfLabelName:    "differencesTest",
					apis.WorkloadTypeLabelName: "web",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Git: &cartov1alpha1.GitSource{
						URL: "example.com",
						Ref: cartov1alpha1.GitRef{
							Branch: "main",
							Tag:    "v1",
						},
					},
				},
				Image: "ubuntu:bionic",
			},
		},
		right: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "change",
				Labels: map[string]string{
					apis.AppPartOfLabelName:    "differencesTest1",
					apis.WorkloadTypeLabelName: "web",
				},
			},
			Spec: cartov1alpha1.WorkloadSpec{
				Source: &cartov1alpha1.Source{
					Git: &cartov1alpha1.GitSource{
						URL: "example.com",
						Ref: cartov1alpha1.GitRef{
							Branch: "main",
							Tag:    "v1",
						},
					},
				},
				Image: "ubuntu:focal",
			},
		},
		want: `
...
  2,  2   |apiVersion: carto.run/v1alpha1
  3,  3   |kind: Workload
  4,  4   |metadata:
  5,  5   |  labels:
  6     - |    app.kubernetes.io/part-of: differencesTest
      6 + |    app.kubernetes.io/part-of: differencesTest1
  7,  7   |    apps.tanzu.vmware.com/workload-type: web
  8,  8   |  name: change
  9,  9   |  namespace: default
 10, 10   |spec:
 11     - |  image: ubuntu:bionic
     11 + |  image: ubuntu:focal
 12, 12   |  source:
 13, 13   |    git:
 14, 14   |      ref:
 15, 15   |        branch: main
...
`,
	}, {
		name:   "no differences",
		scheme: scheme,
		left: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unchanged",
			},
		},
		right: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "unchanged",
			},
		},
		noChange: true,
		want: `
...
`,
	}, {
		name:   "create resource",
		scheme: scheme,
		left:   nil,
		right: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "create",
			},
		},
		want: `
      1 + |---
      2 + |apiVersion: carto.run/v1alpha1
      3 + |kind: Workload
      4 + |metadata:
      5 + |  name: create
      6 + |spec: {}
`,
	}, {
		name:   "delete resource",
		scheme: scheme,
		left: &cartov1alpha1.Workload{
			ObjectMeta: metav1.ObjectMeta{
				Name: "delete",
			},
		},
		right: nil,
		want: `
  1     - |---
  2     - |apiVersion: carto.run/v1alpha1
  3     - |kind: Workload
  4     - |metadata:
  5     - |  name: delete
  6     - |spec: {}
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, noChange, err := printer.ResourceDiff(test.left, test.right, test.scheme)
			if (err != nil) != test.shouldError {
				t.Errorf("ResourceDiff() error = %v, expected %v", err, test.shouldError)
			}
			if noChange != test.noChange {
				t.Errorf("ResourceDiff() noChange = %v, expected %v", noChange, test.noChange)
			}
			if diff := cmp.Diff(strings.TrimPrefix(test.want, "\n"), got); diff != "" {
				t.Errorf("ResourceDiff() (-want, +got) = %v", diff)
			}
		})
	}
}
