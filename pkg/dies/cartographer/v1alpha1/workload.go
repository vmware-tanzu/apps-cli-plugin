package v1alpha1

import (
	diecorev1 "dies.dev/apis/core/v1"
	diemetav1 "dies.dev/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
)

// +die:object=true
type _ = cartov1alpha1.Workload

// +die
type _ = cartov1alpha1.WorkloadSpec

// +die
type _ = cartov1alpha1.WorkloadStatus

func (d *WorkloadSpecDie) EnvDie(name string, fn func(d *diecorev1.EnvVarDie)) *WorkloadSpecDie {
	return d.DieStamp(func(r *cartov1alpha1.WorkloadSpec) {
		for i := range r.Env {
			if name == r.Env[i].Name {
				d := diecorev1.EnvVarBlank.DieImmutable(false).DieFeed(r.Env[i])
				fn(d)
				r.Env[i] = d.DieRelease()
				return
			}
		}

		d := diecorev1.EnvVarBlank.DieImmutable(false).DieFeed(corev1.EnvVar{Name: name})
		fn(d)
		r.Env = append(r.Env, d.DieRelease())
	})
}
func (d *WorkloadStatusDie) ConditionsDie(conditions ...*diemetav1.ConditionDie) *WorkloadStatusDie {
	return d.DieStamp(func(r *cartov1alpha1.WorkloadStatus) {
		r.Conditions = make([]metav1.Condition, len(conditions))
		for i := range conditions {
			r.Conditions[i] = conditions[i].DieRelease()
		}
	})
}

var (
	WorkloadConditionReadyBlank = diemetav1.ConditionBlank.Type(cartov1alpha1.WorkloadConditionReady)
)
