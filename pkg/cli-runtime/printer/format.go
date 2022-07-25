/*
Copyright 2019 VMware, Inc.

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

package printer

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/duration"
)

func TimestampSince(timestamp metav1.Time, now time.Time) string {
	if timestamp.IsZero() {
		return Swarnf("<unknown>")
	}
	return duration.HumanDuration(now.Sub(timestamp.Time))
}

func EmptyString(str string) string {
	if str == "" {
		return Sfaintf("<empty>")
	}
	return str
}

func ConditionStatus(cond *metav1.Condition) string {
	if cond == nil || cond.Status == "" {
		return Swarnf("<unknown>")
	}
	status := string(cond.Status)
	switch status {
	case "True":
		return Ssuccessf(string(cond.Type))
	case "False":
		if cond.Reason == "" {
			// display something if there is no reason
			return Serrorf("not-" + string(cond.Type))
		}
		return Serrorf(cond.Reason)
	default:
		return Sinfof(status)
	}
}

func ColorConditionStatus(condStatus string) string {
	switch condStatus {
	case "True", "true":
		return Ssuccessf(condStatus)
	case "False", "false":
		return Serrorf(condStatus)
	default:
		return Sinfof(condStatus)
	}
}

func Labels(labelMap map[string]string) string {
	return EmptyString(labels.Set(labelMap).String())
}
