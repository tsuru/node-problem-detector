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
	"encoding/json"
	"fmt"
	"k8s.io/node-problem-detector/pkg/types"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	testclock "k8s.io/utils/clock/testing"

	"github.com/stretchr/testify/assert"
)

const (
	testSource = "test"
	testNode   = "test-node"
)

func newFakeProblemClient() *nodeProblemClient {
	return &nodeProblemClient{
		nodeName: testNode,
		// There is no proper fake for *client.Client for now
		// TODO(random-liu): Add test for SetConditions when we have good fake for *client.Client
		clock:     testclock.NewFakeClock(time.Now()),
		recorders: make(map[string]record.EventRecorder),
		nodeRef:   getNodeRef("", testNode),
	}
}

func TestGeneratePatch(t *testing.T) {
	now := time.Now()
	update := []v1.NodeCondition{
		{
			Type:               "TestType1",
			Status:             v1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(now),
			Reason:             "TestReason1",
			Message:            "TestMessage1",
		},
		{
			Type:               "TestType2",
			Status:             v1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(now),
			Reason:             "TestReason2",
			Message:            "TestMessage2",
		},
	}
	raw, err := json.Marshal(&update)
	assert.NoError(t, err)
	expectedPatch := []byte(fmt.Sprintf(`{"status":{"conditions":%s}}`, raw))

	patch, err := generatePatch(update)
	assert.NoError(t, err)
	if string(patch) != string(expectedPatch) {
		t.Errorf("expected patch %q, got %q", expectedPatch, patch)
	}
}

func TestEvent(t *testing.T) {
	fakeRecorder := record.NewFakeRecorder(1)
	client := newFakeProblemClient()
	client.recorders[testSource] = fakeRecorder
	client.Eventf(v1.EventTypeWarning, testSource, "test reason", "test message")
	expected := fmt.Sprintf("%s %s %s", v1.EventTypeWarning, "test reason", "test message")
	got := <-fakeRecorder.Events
	if expected != got {
		t.Errorf("expected event %q, got %q", expected, got)
	}
}

func TestCheckIfTaintAlreadyExists(t *testing.T) {
	cases := []struct {
		node   *v1.Node
		conf   types.TaintConfig
		result bool
	}{
		{&v1.Node{
			Spec: v1.NodeSpec{
				Taints: []v1.Taint{
					{
						Key:    "node-problem-detector/read-only-filesystem",
						Value:  "true",
						Effect: v1.TaintEffectNoSchedule,
					},
				},
			},
		}, types.TaintConfig{
			Enabled: false,
			Key:     "node-problem-detector/read-only-filesystem",
			Value:   "true",
			Effect:  string(v1.TaintEffectNoSchedule),
		},
			true,
		},
		{&v1.Node{
			Spec: v1.NodeSpec{
				Taints: []v1.Taint{
					{
						Key:    "node-problem-detector/read-only-filesystem",
						Value:  "true",
						Effect: v1.TaintEffectNoSchedule,
					},
				},
			},
		}, types.TaintConfig{
			Enabled: false,
			Key:     "node-problem-detector/read-write-filesystem",
			Value:   "true",
			Effect:  string(v1.TaintEffectNoSchedule),
		},
			false,
		},
	}

	for _, tc := range cases {
		exists := CheckIfTaintAlreadyExists(tc.node, tc.conf)

		if tc.result {
			assert.True(t, exists)
		} else {
			assert.False(t, exists)
		}
	}
}
