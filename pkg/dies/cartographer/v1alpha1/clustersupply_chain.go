package v1alpha1

import (
	diemetav1 "dies.dev/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
)

// +die:object=true
type _ = cartov1alpha1.ClusterSupplyChain

// +die
type _ = cartov1alpha1.SupplyChainSpec

// +die
type _ = cartov1alpha1.SupplyChainStatus

func (d *SupplyChainStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *SupplyChainStatusDie {
	return d.DieStamp(func(r *cartov1alpha1.SupplyChainStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

// StatusDie stamps the resource's Status field with a mutable die.
func (d *ClusterSupplyChainDie) StatusDie(fn func(d *SupplyChainStatusDie)) *ClusterSupplyChainDie {
	return d.DieStamp(func(r *cartov1alpha1.ClusterSupplyChain) {
		d := SupplyChainStatusBlank.DieImmutable(false).DieFeed(r.Status)
		fn(d)
		r.Status = d.DieRelease()
	})
}

var (
	ClusterSupplyChainConditionTemplatesReadyBlank = diemetav1.ConditionBlank.Type(cartov1alpha1.SupplyChainTemplatesReady)
	ClusterSupplyChainConditionReadyBlank          = diemetav1.ConditionBlank.Type(cartov1alpha1.SupplyChainReady)
)
