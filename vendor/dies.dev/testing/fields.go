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
	"reflect"

	"k8s.io/apimachinery/pkg/util/sets"
)

func DieFieldDiff(die interface{}) sets.String {
	d := reflect.ValueOf(die).Type()
	ds := sets.NewString()
	for i := 0; i < d.NumMethod(); i++ {
		m := d.Method(i)
		if m.IsExported() {
			ds.Insert(m.Name)
		}
	}

	dr, _ := d.MethodByName("DieRelease")
	r := dr.Type.Out(0)
	rs := sets.NewString()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		if f.IsExported() {
			rs.Insert(f.Name)
		}
	}

	return rs.Difference(ds)
}
