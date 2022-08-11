package v1alpha1

import (
	diemetav1 "dies.dev/apis/meta/v1"

	cartov1alpha1 "github.com/vmware-tanzu/apps-cli-plugin/pkg/apis/cartographer/v1alpha1"
)

var (
	ConditionReadyBlank           = diemetav1.ConditionBlank.Type(cartov1alpha1.ConditionReady)
	ConditionHealthyBlank         = diemetav1.ConditionBlank.Type(cartov1alpha1.ResourcesHealthy)
	ConditionResourceReadyBlank   = diemetav1.ConditionBlank.Type(cartov1alpha1.ConditionResourceReady)
	ConditionResourceHealthyBlank = diemetav1.ConditionBlank.Type(cartov1alpha1.ConditionResourceHealthy)
)

func CreateConditionReadyTrue(reason, message string) *diemetav1.ConditionDie {
	return ConditionReadyBlank.
		True().
		Reason(reason).
		Message(message)
}

func CreateConditionHealthyTrue(reason, message string) *diemetav1.ConditionDie {
	return ConditionHealthyBlank.
		True().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceReadyTrue(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceReadyBlank.
		True().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceHealthyTrue(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceHealthyBlank.
		True().
		Reason(reason).
		Message(message)
}

func CreateConditionReadyFalse(reason, message string) *diemetav1.ConditionDie {
	return ConditionReadyBlank.
		False().
		Reason(reason).
		Message(message)
}

func CreateConditionHealthyFalse(reason, message string) *diemetav1.ConditionDie {
	return ConditionHealthyBlank.
		False().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceReadyFalse(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceReadyBlank.
		False().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceHealthyFalse(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceHealthyBlank.
		False().
		Reason(reason).
		Message(message)
}

func CreateConditionReadyUnknown(reason, message string) *diemetav1.ConditionDie {
	return ConditionReadyBlank.
		Unknown().
		Reason(reason).
		Message(message)
}

func CreateConditionHealthyUnknown(reason, message string) *diemetav1.ConditionDie {
	return ConditionHealthyBlank.
		Unknown().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceReadyUnknown(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceReadyBlank.
		Unknown().
		Reason(reason).
		Message(message)
}

func CreateConditionResourceHealthyUnknown(reason, message string) *diemetav1.ConditionDie {
	return ConditionResourceHealthyBlank.
		Unknown().
		Reason(reason).
		Message(message)
}

func (d *RealizedResourceDie) ConditionsHealthyReadyTrueDie() *RealizedResourceDie {
	return d.ConditionsDie(CreateConditionReadyTrue("", ""), CreateConditionHealthyTrue("", ""))
}

func (d *RealizedResourceDie) ConditionsResourceHealthyReadyTrueDie() *RealizedResourceDie {
	return d.ConditionsDie(CreateConditionResourceReadyTrue("", ""), CreateConditionResourceHealthyTrue("", ""))
}
