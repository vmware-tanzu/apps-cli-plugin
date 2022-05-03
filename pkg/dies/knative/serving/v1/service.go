package v1

import (
	knativeservingv1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/knative/serving/v1"
)

// +die:object=true
type _ = knativeservingv1.Service

// +die
type _ = knativeservingv1.ServiceStatus
