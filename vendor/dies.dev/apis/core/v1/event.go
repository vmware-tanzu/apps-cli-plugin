/*
Copyright 2021 the original author or authors.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// +die:object=true
type _ = corev1.Event

func (d *EventDie) InvolvedObjectDie(fn func(d *ObjectReferenceDie)) *EventDie {
	return d.DieStamp(func(r *corev1.Event) {
		d := ObjectReferenceBlank.DieImmutable(false).DieFeed(r.InvolvedObject)
		fn(d)
		r.InvolvedObject = d.DieRelease()
	})
}

func (d *EventDie) SourceDie(fn func(d *EventSourceDie)) *EventDie {
	return d.DieStamp(func(r *corev1.Event) {
		d := EventSourceBlank.DieImmutable(false).DieFeed(r.Source)
		fn(d)
		r.Source = d.DieRelease()
	})
}

func (d *EventDie) SeriesDie(fn func(d *EventSeriesDie)) *EventDie {
	return d.DieStamp(func(r *corev1.Event) {
		d := EventSeriesBlank.DieImmutable(false).DieFeedPtr(r.Series)
		fn(d)
		r.Series = d.DieReleasePtr()
	})
}

func (d *EventDie) RelatedDie(fn func(d *ObjectReferenceDie)) *EventDie {
	return d.DieStamp(func(r *corev1.Event) {
		d := ObjectReferenceBlank.DieImmutable(false).DieFeedPtr(r.Related)
		fn(d)
		r.Related = d.DieReleasePtr()
	})
}

// +die
type _ = corev1.EventSource

// +die
type _ = corev1.EventSeries
