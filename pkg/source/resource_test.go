package source_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	k8sscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	clitesting "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/testing"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/source"
)

func TestWorkloadOptionsFetchResourceObject(t *testing.T) {
	defaultNamespace := "default"
	workloadName := "my-workload"
	pod1Die := diecorev1.PodBlank.
		MetadataDie(func(d *diemetav1.ObjectMetaDie) {
			d.Name("pod1")
			d.Namespace(defaultNamespace)
			d.AddLabel(cartov1alpha1.WorkloadLabelName, workloadName)
		}).Kind("pod")

	scheme := runtime.NewScheme()
	c := cli.NewDefaultConfig("test", scheme)
	c.Client = clitesting.NewFakeCliClient(clitesting.NewFakeClient(scheme))

	tests := []struct {
		name           string
		args           []string
		givenWorkload  *cartov1alpha1.Workload
		shouldError    bool
		expectedOutput string
		withReactors   []clitesting.ReactionFunc
	}{
		{
			name: "Fetch Resource Object successfully",
			args: []string{flags.LabelFlagName, "NEW=value", flags.YesFlagName},
			givenWorkload: &cartov1alpha1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: defaultNamespace,
					Name:      workloadName,
					Labels: map[string]string{
						"FOO": "bar",
					},
				},
				Spec: cartov1alpha1.WorkloadSpec{
					Image: "ubuntu",
				},
			},
			shouldError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = cartov1alpha1.AddToScheme(scheme)
			c := cli.NewDefaultConfig("test", scheme)
			output := &bytes.Buffer{}
			c.Stdout = output
			c.Stderr = output
			fakeClient := clitesting.NewFakeClient(scheme, test.givenWorkload)

			for i := range test.withReactors {
				// in reverse order since we prepend
				reactor := test.withReactors[len(test.withReactors)-1-i]
				fakeClient.PrependReactor("*", "*", reactor)
			}

			c.Client = clitesting.NewFakeCliClient(fakeClient)

			cmd := &cobra.Command{}
			ctx := cli.WithCommand(context.Background(), cmd)
			c.Builder = resource.NewFakeBuilder(
				func(version schema.GroupVersion) (resource.RESTClient, error) {
					codec := k8sscheme.Codecs.LegacyCodec(scheme.PrioritizedVersionsAllGroups()...)
					UnstructuredClient := &fake.RESTClient{
						NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
						Resp:                 &http.Response{StatusCode: http.StatusOK, Header: clitesting.DefaultHeader(), Body: clitesting.PodV1TableObjBody(codec, []client.Object{pod1Die})},
					}
					return UnstructuredClient, nil

				},
				c.ToRESTMapper,
				func() (restmapper.CategoryExpander, error) {
					return resource.FakeCategoryExpander, nil
				},
			)

			currentWorkload := &cartov1alpha1.Workload{}
			err := c.Get(ctx, client.ObjectKey{Namespace: defaultNamespace, Name: workloadName}, currentWorkload)

			if err != nil {
				t.Errorf("Update() errored %v", err)
			}

			opts := &commands.WorkloadOptions{}
			opts.DefineFlags(ctx, c, cmd)
			cmd.ParseFlags(test.args)
			arg := []string{"pods"}
			workload := currentWorkload.DeepCopy()
			opts.ApplyOptionsToWorkload(ctx, workload)
			_, err = source.FetchResourceObject(c, workload, arg)

			if err != nil && !test.shouldError {
				t.Errorf("Update() errored %v", err)
			}
			if err == nil && test.shouldError {
				t.Errorf("Update() expected error")
			}
			if test.shouldError {
				return
			}
		})
	}
}
