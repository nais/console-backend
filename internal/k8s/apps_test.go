package k8s

import (
	"testing"

	"github.com/nais/console-backend/internal/graph/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testCase struct {
	appCondition     AppCondition
	instanceStates   []model.InstanceState
	expectedState    model.AppState
	expectedMessages []string
}

func TestSetStatus(t *testing.T) {
	testCases := []testCase{
		{
			appCondition:     AppConditionRolloutComplete,
			instanceStates:   []model.InstanceState{model.InstanceStateRunning},
			expectedState:    model.AppStateNais,
			expectedMessages: nil,
		},
		{
			appCondition:     AppConditionRolloutComplete,
			instanceStates:   []model.InstanceState{model.InstanceStateFailing},
			expectedState:    model.AppStateFailing,
			expectedMessages: []string{"No running instances"},
		},
		{
			appCondition:     AppConditionFailedSynchronization,
			instanceStates:   []model.InstanceState{model.InstanceStateRunning},
			expectedState:    model.AppStateNotnais,
			expectedMessages: []string{"Invalid nais.yaml"},
		},
		{
			appCondition:     AppConditionFailedSynchronization,
			instanceStates:   []model.InstanceState{model.InstanceStateFailing},
			expectedState:    model.AppStateFailing,
			expectedMessages: []string{"No running instances", "Invalid nais.yaml"},
		},
		{
			appCondition:     AppConditionSynchronized,
			instanceStates:   []model.InstanceState{model.InstanceStateRunning, model.InstanceStateFailing},
			expectedState:    model.AppStateNotnais,
			expectedMessages: []string{"New instances failing"},
		},
		{
			appCondition:     AppConditionSynchronized,
			instanceStates:   []model.InstanceState{model.InstanceStateFailing, model.InstanceStateFailing},
			expectedState:    model.AppStateFailing,
			expectedMessages: []string{"New instances failing", "No running instances"},
		},
	}

	for _, tc := range testCases {
		app := &model.App{}
		setStatus(app, []metav1.Condition{{Status: metav1.ConditionTrue, Reason: string(tc.appCondition)}}, asInstances(tc.instanceStates))

		if app.State != tc.expectedState {
			t.Errorf("got: %v, want: %v", app.State, tc.expectedState)
		}
		if !hasMessage(app.Messages, tc.expectedMessages) {
			t.Errorf("got: %v, want: %v", app.Messages, tc.expectedMessages)
		}
	}
}

func hasMessage(messages []string, expectedMessages []string) bool {
	if len(messages) != len(expectedMessages) {
		return false
	}

	for _, msg := range expectedMessages {
		if !contains(messages, msg) {
			return false
		}
	}
	return true
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func asInstances(states []model.InstanceState) []*model.Instance {
	var ret []*model.Instance
	for _, state := range states {
		ret = append(ret, &model.Instance{State: state})
	}
	return ret
}
