package source

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
)

func FetchResourceObject(c *cli.Config, workload *cartov1alpha1.Workload, args []string) (runtime.Object, error) {
	r := c.Builder.Unstructured().
		NamespaceParam(workload.Namespace).
		LabelSelectorParam(fmt.Sprintf("%s%s%s", cartov1alpha1.WorkloadLabelName, "=", workload.Name)).
		ResourceTypeOrNameArgs(true, args...).
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) {
			req.SetHeader("Accept", strings.Join([]string{
				fmt.Sprintf(runtime.ContentTypeJSON+";as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
				runtime.ContentTypeJSON,
			}, ","))
		}).
		Do()
	infos, err := r.Infos()
	if err != nil {
		return nil, fmt.Errorf("failed to list pods:\n  %s", err)
	}
	return decodeIntoTable(infos)
}

func decodeIntoTable(info []*resource.Info) (runtime.Object, error) {
	if len(info) == 1 {
		obj := info[0].Object
		event, isEvent := obj.(*metav1.WatchEvent)
		if isEvent {
			obj = event.Object.Object
		}
		if !recognizedTableVersions[obj.GetObjectKind().GroupVersionKind()] {
			return nil, fmt.Errorf("attempt to decode non-Table object")
		}

		unstr, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("attempt to decode non-Unstructured object")
		}
		table := &metav1.Table{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstr.Object, table); err != nil {
			return nil, err
		}
		if len(table.Rows) == 0 {
			return nil, nil
		}
		for i := range table.Rows {
			row := &table.Rows[i]
			if row.Object.Raw == nil || row.Object.Object != nil {
				continue
			}
			var converted runtime.Object
			var err error

			converted, err = runtime.Decode(unstructured.UnstructuredJSONScheme, row.Object.Raw)
			if err != nil {
				return nil, err
			}

			row.Object.Object = converted
		}

		return table, nil
	}
	return nil, nil
}

// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through

var recognizedTableVersions = map[schema.GroupVersionKind]bool{
	metav1beta1.SchemeGroupVersion.WithKind("Table"): true,
	metav1.SchemeGroupVersion.WithKind("Table"):      true,
}
