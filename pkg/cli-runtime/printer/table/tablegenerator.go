/*
Copyright 2019 The Kubernetes Authors.

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

// repackaged from https://github.com/kubernetes/kubernetes/tree/v1.15.0-beta.0/pkg/printers
// TODO remove once we can depend directly on k8s 1.15

package table

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// TableGenerator - an interface for generating metav1beta1.Table provided a runtime.Object
type TableGenerator interface {
	GenerateTable(obj runtime.Object, options PrintOptions) (*metav1beta1.Table, error)
}

// PrintHandler - interface to handle printing provided an array of metav1beta1.TableColumnDefinition
type PrintHandler interface {
	TableHandler(columns []metav1beta1.TableColumnDefinition, printFunc interface{}) error
}

type handlerEntry struct {
	columnDefinitions []metav1beta1.TableColumnDefinition
	printFunc         reflect.Value
	args              []reflect.Value
}

// HumanReadablePrinter is an implementation of ResourcePrinter which attempts to provide
// more elegant output. It is not threadsafe, but you may call PrintObj repeatedly; headers
// will only be printed if the object type changes. This makes it useful for printing items
// received from watches.
type HumanReadablePrinter struct {
	handlerMap  map[reflect.Type]*handlerEntry
	options     PrintOptions
	lastType    interface{}
	lastColumns []metav1beta1.TableColumnDefinition
}

var _ TableGenerator = &HumanReadablePrinter{}
var _ PrintHandler = &HumanReadablePrinter{}

// NewTableGenerator creates a HumanReadablePrinter suitable for calling GenerateTable().
func NewTableGenerator() *HumanReadablePrinter {
	return &HumanReadablePrinter{
		handlerMap: make(map[reflect.Type]*handlerEntry),
	}
}

// With method - accepts a list of builder functions that modify HumanReadablePrinter
func (h *HumanReadablePrinter) With(fns ...func(PrintHandler)) *HumanReadablePrinter {
	for _, fn := range fns {
		fn(h)
	}
	return h
}

// GenerateTable returns a table for the provided object, using the printer registered for that type. It returns
// a table that includes all of the information requested by options, but will not remove rows or columns. The
// caller is responsible for applying rules related to filtering rows or columns.
func (h *HumanReadablePrinter) GenerateTable(obj runtime.Object, options PrintOptions) (*metav1beta1.Table, error) {
	t := reflect.TypeOf(obj)
	handler, ok := h.handlerMap[t]
	if !ok {
		return nil, fmt.Errorf("no table handler registered for this type %v", t)
	}

	args := []reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(options)}
	results := handler.printFunc.Call(args)
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	var columns []metav1beta1.TableColumnDefinition
	if !options.NoHeaders {
		columns = handler.columnDefinitions
		if !options.Wide {
			columns = make([]metav1beta1.TableColumnDefinition, 0, len(handler.columnDefinitions))
			for i := range handler.columnDefinitions {
				if handler.columnDefinitions[i].Priority != 0 {
					continue
				}
				columns = append(columns, handler.columnDefinitions[i])
			}
		}
	}
	table := &metav1beta1.Table{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "",
		},
		ColumnDefinitions: columns,
		Rows:              results[0].Interface().([]metav1beta1.TableRow),
	}
	if m, err := meta.ListAccessor(obj); err == nil {
		table.ResourceVersion = m.GetResourceVersion()
		table.SelfLink = m.GetSelfLink()
		table.Continue = m.GetContinue()
		table.RemainingItemCount = m.GetRemainingItemCount()
	} else {
		if m, err := meta.CommonAccessor(obj); err == nil {
			table.ResourceVersion = m.GetResourceVersion()
			table.SelfLink = m.GetSelfLink()
		}
	}
	if err := decorateTable(table, options); err != nil {
		return nil, err
	}
	return table, nil
}

// TableHandler adds a print handler with a given set of columns to HumanReadablePrinter instance.
// See ValidateRowPrintHandlerFunc for required method signature.
func (h *HumanReadablePrinter) TableHandler(columnDefinitions []metav1beta1.TableColumnDefinition, printFunc interface{}) error {
	printFuncValue := reflect.ValueOf(printFunc)
	if err := ValidateRowPrintHandlerFunc(printFuncValue); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to register print function: %v", err))
		return err
	}
	entry := &handlerEntry{
		columnDefinitions: columnDefinitions,
		printFunc:         printFuncValue,
	}

	objType := printFuncValue.Type().In(0)
	if _, ok := h.handlerMap[objType]; ok {
		err := fmt.Errorf("registered duplicate printer for %v", objType)
		utilruntime.HandleError(err)
		return err
	}
	h.handlerMap[objType] = entry
	return nil
}

// ValidateRowPrintHandlerFunc validates print handler signature.
// printFunc is the function that will be called to print an object.
// It must be of the following type:
//
//	func printFunc(object ObjectType, options PrintOptions) ([]metav1beta1.TableRow, error)
//
// where ObjectType is the type of the object that will be printed, and the first
// return value is an array of rows, with each row containing a number of cells that
// match the number of columns defined for that printer function.
func ValidateRowPrintHandlerFunc(printFunc reflect.Value) error {
	if printFunc.Kind() != reflect.Func {
		return fmt.Errorf("invalid print handler. %#v is not a function", printFunc)
	}
	funcType := printFunc.Type()
	if funcType.NumIn() != 2 || funcType.NumOut() != 2 {
		return fmt.Errorf("invalid print handler." +
			"Must accept 2 parameters and return 2 value.")
	}
	if funcType.In(1) != reflect.TypeOf((*PrintOptions)(nil)).Elem() ||
		funcType.Out(0) != reflect.TypeOf((*[]metav1beta1.TableRow)(nil)).Elem() ||
		funcType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("invalid print handler. The expected signature is: "+
			"func handler(obj %v, options PrintOptions) ([]metav1beta1.TableRow, error)", funcType.In(0))
	}
	return nil
}
