// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "strings"

// StringToTarget converts string to Target type
func StringToTarget(target string) Target {
	if target == string(targetK8s) || target == string(TargetK8s) {
		return TargetK8s
	} else if target == string(targetTMC) || target == string(TargetTMC) {
		return TargetTMC
	} else if target == string(TargetGlobal) {
		return TargetGlobal
	} else if target == string(TargetUnknown) {
		return TargetUnknown
	}
	return TargetUnknown
}

// IsValidTarget validates the target string specified is valid or not
// TargetGlobal and TargetUnknown are special targets and hence this function
// provide flexibility additional arguments to allow them based on the requirement
func IsValidTarget(target string, allowGlobal, allowUnknown bool) bool {
	return target == string(targetK8s) ||
		target == string(TargetK8s) ||
		target == string(targetTMC) ||
		target == string(TargetTMC) ||
		(allowGlobal && target == string(TargetGlobal)) ||
		(allowUnknown && target == string(TargetUnknown))
}

// StringToContextType converts string to ContextType
func StringToContextType(contextType string) ContextType {
	contextType = strings.ToLower(contextType)
	if contextType == string(contextTypeK8s) || contextType == string(ContextTypeK8s) {
		return ContextTypeK8s
	} else if contextType == string(contextTypeTMC) || contextType == string(ContextTypeTMC) {
		return ContextTypeTMC
	} else if contextType == string(ContextTypeTanzu) {
		return ContextTypeTanzu
	}
	return ""
}

// IsValidContextType validates the contextType string specified is valid or not
func IsValidContextType(contextType string) bool {
	ct := StringToContextType(contextType)
	if ct == "" && contextType != "" {
		return false
	}
	return true
}
