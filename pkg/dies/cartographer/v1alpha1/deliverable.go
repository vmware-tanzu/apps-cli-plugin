package v1alpha1

import (
	diemetav1 "dies.dev/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
)

// +die:object=true
type _ = cartov1alpha1.Deliverable

// +die
type _ = cartov1alpha1.DeliverableSpec

// +die
type _ = cartov1alpha1.DeliverableStatus

func (d *DeliverableStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *DeliverableStatusDie {
	return d.DieStamp(func(r *cartov1alpha1.DeliverableStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

func (d *DeliverableDie) ConditionsHealthyReadyTrueDie() *DeliverableDie {
	d.StatusDie(func(d *DeliverableStatusDie) {
		d.ConditionsDie(CreateConditionReadyTrue("", ""), CreateConditionHealthyTrue("", ""))
	})
	return d
}

func (d *DeliverableStatusDie) ConditionsResourceReadyHealthyTrueDie() *DeliverableStatusDie {
	return d.Resources(RealizedResourceBlank.ConditionsResourceHealthyReadyTrueDie().DieRelease())
}
