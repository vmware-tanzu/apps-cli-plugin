/*
Copyright 2021 VMware, Inc.

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

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
	cli "github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/cli-runtime/validation"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/commands/resource"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/completion"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/flags"
	"github.com/vmware-tanzu/apps-cli-plugin/pkg/printer"
)

type WorkloadGetOptions struct {
	Namespace string
	Name      string

	Export bool
	Output string
}

var (
	_ validation.Validatable = (*WorkloadGetOptions)(nil)
	_ cli.Executable         = (*WorkloadGetOptions)(nil)
)

func (opts *WorkloadGetOptions) Validate(ctx context.Context) validation.FieldErrors {
	errs := validation.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(validation.ErrMissingField(flags.NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(validation.ErrMissingField(cli.NameArgumentName))
	}

	if opts.Output != "" {
		errs = errs.Also(validation.Enum(opts.Output, flags.OutputFlagName, []string{printer.OutputFormatJson, printer.OutputFormatYaml, printer.OutputFormatYml}))
	}

	return errs
}
func (opts *WorkloadGetOptions) transformRequests(req *rest.Request) {

	req.SetHeader("Accept", strings.Join([]string{
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
		fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
		"application/json",
	}, ","))

	// if sorting, ensure we receive the full object in order to introspect its fields via jsonpath
	if true {
		req.Param("includeObject", "Object")
	}
}
func (opts *WorkloadGetOptions) Exec(ctx context.Context, c *cli.Config) error {
	workload := &cartov1alpha1.Workload{}
	err := c.Get(ctx, client.ObjectKey{Namespace: opts.Namespace, Name: opts.Name}, workload)
	if err != nil {
		if apierrs.IsNotFound(err) {
			nsGet := &corev1.Namespace{}
			if getErr := c.Get(ctx, types.NamespacedName{Name: opts.Namespace}, nsGet); getErr != nil && apierrs.IsNotFound(getErr) {
				c.Eprintf("%s %s\n", printer.Serrorf("Error:"), fmt.Sprintf("namespace %q not found, it may not exist or user does not have permissions to read it.", opts.Namespace))
				return cli.SilenceError(getErr)
			}
			c.Errorf("Workload %q not found\n", fmt.Sprintf("%s/%s", opts.Namespace, opts.Name))
			return cli.SilenceError(err)
		}

		return err
	}

	if opts.Export {
		var format printer.OutputFormat
		if opts.Output == "" {
			format = printer.OutputFormat(printer.OutputFormatYaml)
		} else {
			format = printer.OutputFormat(opts.Output)
		}

		export, err := printer.ExportResource(workload, format, c.Scheme)
		if err != nil {
			c.Eprintf("%s %s\n", printer.Serrorf("Failed to export workload:"), err)
			return cli.SilenceError(err)
		}
		c.Printf("%s\n", export)
		return nil
	}

	if opts.Output != "" {
		export, err := printer.OutputResource(workload, printer.OutputFormat(opts.Output), c.Scheme)
		if err != nil {
			c.Eprintf("%s %s\n", printer.Serrorf("Failed to output workload:"), err)
			return cli.SilenceError(err)
		}

		c.Printf("%s\n", export)
		return nil
	}

	workloadStatusReadyCond := printer.FindCondition(workload.Status.Conditions, cartov1alpha1.WorkloadConditionReady)
	c.Printf(printer.ResourceStatus(workload.Name, workloadStatusReadyCond))
	//print workload details
	c.Boldf("Overview\n")
	if err := printer.WorkloadOverviewPrinter(c.Stdout, workload); err != nil {
		return err
	}
	c.Printf("\n")
	// Print workload source
	if workload.Spec.Image != "" || workload.Spec.Source != nil {
		c.Boldf("Source\n")

		if workload.Spec.Image != "" {
			if err := printer.WorkloadSourceImagePrinter(c.Stdout, workload); err != nil {
				return err
			}
		}

		if workload.Spec.Source != nil {
			if workload.Spec.Source.Image != "" {
				if err := printer.WorkloadLocalSourceImagePrinter(c.Stdout, workload); err != nil {
					return err
				}
			}
			if workload.Spec.Source.Git != nil {
				if err := printer.WorkloadSourceGitPrinter(c.Stdout, workload); err != nil {
					return err
				}
			}
		}
		c.Printf("\n")
	}

	// Print workload supply chain
	if workload.Status.SupplyChainRef == (cartov1alpha1.ObjectReference{}) && len(workload.Status.Conditions) == 0 {
		c.Infof("Supply Chain reference not found.\n")
	} else {
		c.Boldf("Supply Chain\n")

		if err := printer.WorkloadSupplyChainInfoPrinter(c.Stdout, workload); err != nil {
			return err
		}
	}

	// Print workload resources
	c.Printf("\n")
	if len(workload.Status.Resources) == 0 {
		c.Infof(printer.AddPaddingStart("Supply Chain resources not found.\n"))
	} else {
		if err := printer.WorkloadResourcesPrinter(c.Stdout, workload); err != nil {
			return err
		}
	}

	// Deliverable
	c.Printf("\n")
	c.Boldf("Delivery\n")
	// Print workload deliverable resources
	wldDeliverable := getWorkloadResourceByKind(workload, cartov1alpha1.DeliverableKind)
	var deliverableStatusReadyCond *metav1.Condition
	notFoundMsg := printer.AddPaddingStart("Delivery resources not found.\n")
	deliverable := &cartov1alpha1.Deliverable{}
	if wldDeliverable != nil {
		if err := c.Get(ctx, client.ObjectKey{Namespace: wldDeliverable.StampedRef.Namespace, Name: wldDeliverable.StampedRef.Name}, deliverable); err != nil {
			c.Printf("\n")
			c.Infof(notFoundMsg)
		} else if deliverable != nil {
			deliverableStatusReadyCond = printer.FindCondition(deliverable.Status.Conditions, cartov1alpha1.ConditionReady)
			if err := printer.DeliveryInfoPrinter(c.Stdout, deliverable); err != nil {
				return err
			}
			c.Printf("\n")
			if len(deliverable.Status.Resources) == 0 {
				c.Infof(notFoundMsg)
			} else if err := printer.DeliverableResourcesPrinter(c.Stdout, deliverable); err != nil {
				return err
			}
		}
	} else {
		c.Printf("\n")
		c.Infof(notFoundMsg)
	}

	// Print workload issues
	c.Printf("\n")
	c.Boldf("Messages\n")
	if areAllResourcesReady(workloadStatusReadyCond, deliverableStatusReadyCond) {
		c.Infof(printer.AddPaddingStart("No messages found.\n"))
	} else {
		if err := printer.WorkloadIssuesPrinter(c.Stdout, workload); err != nil {
			return err
		}
		if err := printer.DeliverableIssuesPrinter(c.Stdout, deliverable); err != nil {
			return err
		}
	}

	if len(workload.Spec.ServiceClaims) > 0 {
		c.Printf("\n")
		c.Boldf("Services\n")
		if err := cartov1alpha1.WorkloadServiceClaimPrinter(c.Stdout, workload); err != nil {
			return err
		}
	}
	arg := []string{"Pod"}
	bldr := resource.NewBuilderFromConf(c)
	r := bldr.Unstructured().
		NamespaceParam(workload.Namespace).DefaultNamespace().AllNamespaces(false).
		// FilenameParam(false, nil).
		LabelSelectorParam(fmt.Sprintf("%s%s%s", cartov1alpha1.WorkloadLabelName, "=", workload.Name)).
		FieldSelectorParam("").
		ResourceTypeOrNameArgs(true, arg...).
		ContinueOnError().
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) {
			req.SetHeader("Accept", strings.Join([]string{
				fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1.SchemeGroupVersion.Version, metav1.GroupName),
				fmt.Sprintf("application/json;as=Table;v=%s;g=%s", metav1beta1.SchemeGroupVersion.Version, metav1beta1.GroupName),
				"application/json",
			}, ","))

			// if sorting, ensure we receive the full object in order to introspect its fields via jsonpath
			if false {
				req.Param("includeObject", "Object")
			}
		}).
		Do()
	infos, _ := r.Infos()
	var info *resource.Info
	// var printer printers.ResourcePrinter
	for ix := range infos {
		// var mapping *meta.RESTMapping
		c.Printf("\n")
		c.Boldf("Pods\n")
		info = infos[ix]
		printer.PodTablePrintery(c, info)
	}

	ksvcs := &knativeservingv1.ServiceList{}
	_ = c.List(ctx, ksvcs, client.InNamespace(workload.Namespace), client.MatchingLabels{cartov1alpha1.WorkloadLabelName: workload.Name})
	if len(ksvcs.Items) > 0 {
		ksvcs = ksvcs.DeepCopy()
		printer.SortByNamespaceAndName(ksvcs.Items)
		c.Printf("\n")
		c.Boldf("Knative Services\n")
		if err := printer.KnativeServicePrinter(c, ksvcs); err != nil {
			return err
		}
	}

	c.Printf("\n")
	if workload.Namespace != c.Client.DefaultNamespace() {
		c.Infof("To see logs: \"tanzu apps workload tail %s %s %s\"\n", workload.Name, flags.NamespaceFlagName, workload.Namespace)
	} else {
		c.Infof("To see logs: \"tanzu apps workload tail %s\"\n", workload.Name)
	}
	c.Printf("\n")

	return nil
}

func NewWorkloadGetCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &WorkloadGetOptions{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details from a workload",
		Long:  strings.TrimSpace(`Get details from a workload`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s workload get my-workload", c.Name),
		}, "\n"),
		PreRunE:           cli.ValidateE(ctx, opts),
		RunE:              cli.ExecE(ctx, c, opts),
		ValidArgsFunction: completion.SuggestWorkloadNames(ctx, c),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(ctx, cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.Export, cli.StripDash(flags.ExportFlagName), false, "export workload in yaml format")
	cmd.Flags().StringVarP(&opts.Output, cli.StripDash(flags.OutputFlagName), "o", "", "output the Workload formatted. Supported formats: \"json\", \"yaml\", \"yml\"")

	return cmd
}

func getWorkloadResourceByKind(workload *cartov1alpha1.Workload, kind string) *cartov1alpha1.RealizedResource {
	for _, resource := range workload.Status.Resources {
		if resource.StampedRef != nil && resource.StampedRef.Kind == kind {
			return &resource
		}
	}
	return nil
}

func areAllResourcesReady(resourcesConditions ...*metav1.Condition) bool {
	for _, condition := range resourcesConditions {
		if ready := condition == nil || (condition.Status == metav1.ConditionTrue || condition.Message == ""); !ready {
			return false
		}
	}
	return true
}
