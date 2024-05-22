/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package problemclient

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"k8s.io/node-problem-detector/pkg/types"

	v1 "k8s.io/api/core/v1"
)

// FakeProblemClient is a fake problem client for debug.
type FakeProblemClient struct {
	sync.Mutex
	conditions map[v1.NodeConditionType]v1.NodeCondition
	errors     map[string]error
}

// NewFakeProblemClient creates a new fake problem client.
func NewFakeProblemClient() *FakeProblemClient {
	return &FakeProblemClient{
		conditions: make(map[v1.NodeConditionType]v1.NodeCondition),
		errors:     make(map[string]error),
	}
}

// InjectError injects error to specific function.
func (f *FakeProblemClient) InjectError(fun string, err error) {
	f.Lock()
	defer f.Unlock()
	f.errors[fun] = err
}

// AssertConditions asserts that the internal conditions in fake problem client should match
// the expected conditions.
func (f *FakeProblemClient) AssertConditions(expected []v1.NodeCondition) error {
	conditions := map[v1.NodeConditionType]v1.NodeCondition{}
	for _, condition := range expected {
		conditions[condition.Type] = condition
	}
	if !reflect.DeepEqual(conditions, f.conditions) {
		return fmt.Errorf("expected %+v, got %+v", conditions, f.conditions)
	}
	return nil
}

// SetConditions is a fake mimic of SetConditions, it only update the internal condition cache.
func (f *FakeProblemClient) SetConditions(ctx context.Context, conditions []v1.NodeCondition) error {
	f.Lock()
	defer f.Unlock()
	if err, ok := f.errors["SetConditions"]; ok {
		return err
	}
	for _, condition := range conditions {
		f.conditions[condition.Type] = condition
	}
	return nil
}

// TaintNode taints the node if tainting is enabled and problem occurred
func (f *FakeProblemClient) TaintNode(ctx context.Context, condition types.Condition) error {
	return fmt.Errorf("TaintNode() not implemented")
}

// UntaintNode removes taint from node if tainting is enabled and problem resolved
func (f *FakeProblemClient) UntaintNode(ctx context.Context, condition types.Condition) error {
	return fmt.Errorf("UntaintNode() not implemented")
}

// GetConditions is a fake mimic of GetConditions, it returns the conditions cached internally.
func (f *FakeProblemClient) GetConditions(ctx context.Context, types []v1.NodeConditionType) ([]*v1.NodeCondition, error) {
	f.Lock()
	defer f.Unlock()
	if err, ok := f.errors["GetConditions"]; ok {
		return nil, err
	}
	conditions := []*v1.NodeCondition{}
	for _, t := range types {
		condition, ok := f.conditions[t]
		if ok {
			conditions = append(conditions, &condition)
		}
	}
	return conditions, nil
}

// Eventf does nothing now.
func (f *FakeProblemClient) Eventf(eventType string, source, reason, messageFmt string, args ...interface{}) {
}

func (f *FakeProblemClient) GetNode(ctx context.Context) (*v1.Node, error) {
	return nil, fmt.Errorf("GetNode() not implemented")
}
