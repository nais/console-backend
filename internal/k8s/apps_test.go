package k8s

import (
	"fmt"
	"testing"

	"github.com/nais/console-backend/internal/graph/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testCase struct {
	name           string
	appCondition   AppCondition
	instanceStates []model.InstanceState
	image          string
	ingresses      []string
	expectedState  model.State
	expectedErrors []model.ErrorType
}

func TestSetStatus(t *testing.T) {
	testCases := []testCase{
		{
			name:           "app is rolloutcomplete and has running instances",
			appCondition:   AppConditionRolloutComplete,
			instanceStates: []model.InstanceState{model.InstanceStateRunning},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateNais,
			expectedErrors: []model.ErrorType{},
		},
		{
			name:           "app is rolloutcomplete and has failing instances",
			appCondition:   AppConditionRolloutComplete,
			instanceStates: []model.InstanceState{model.InstanceStateFailing},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateFailing,
			expectedErrors: []model.ErrorType{model.ErrorTypeNoRunningInstances},
		},
		{
			name:           "app failed synchronization and has running instances",
			appCondition:   AppConditionFailedSynchronization,
			instanceStates: []model.InstanceState{model.InstanceStateRunning},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateNotnais,
			expectedErrors: []model.ErrorType{model.ErrorTypeInvalidNaisYaml},
		},
		{
			name:           "app failed synchronization and has failing instances",
			appCondition:   AppConditionFailedSynchronization,
			instanceStates: []model.InstanceState{model.InstanceStateFailing},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateFailing,
			expectedErrors: []model.ErrorType{model.ErrorTypeNoRunningInstances, model.ErrorTypeInvalidNaisYaml},
		},
		{
			name:           "app is synchronized and has running and failing instances",
			appCondition:   AppConditionSynchronized,
			instanceStates: []model.InstanceState{model.InstanceStateRunning, model.InstanceStateFailing},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateNotnais,
			expectedErrors: []model.ErrorType{model.ErrorTypeNewInstancesFailing},
		},
		{
			name:           "app is synchronized and has multiple failing instances",
			appCondition:   AppConditionSynchronized,
			instanceStates: []model.InstanceState{model.InstanceStateFailing, model.InstanceStateFailing},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateFailing,
			expectedErrors: []model.ErrorType{model.ErrorTypeNewInstancesFailing, model.ErrorTypeNoRunningInstances},
		},
		{
			name:           "app is rolloutcomplete and has no instances",
			appCondition:   AppConditionRolloutComplete,
			instanceStates: []model.InstanceState{},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateFailing,
			expectedErrors: []model.ErrorType{model.ErrorTypeNoRunningInstances},
		},
		{
			name:           "app image is from deprecated registry",
			appCondition:   AppConditionRolloutComplete,
			instanceStates: []model.InstanceState{model.InstanceStateRunning, model.InstanceStateRunning},
			image:          "docker.pkg.github.com/nais/myapp/myapp:1.0.0",
			ingresses:      []string{"myapp.nav.cloud.nais.io"},
			expectedState:  model.StateNotnais,
			expectedErrors: []model.ErrorType{model.ErrorTypeDeprecatedRegistry},
		},
		{
			name:           "app has deprecated ingress",
			appCondition:   AppConditionRolloutComplete,
			instanceStates: []model.InstanceState{model.InstanceStateRunning, model.InstanceStateRunning},
			image:          "europe-north1-docker.pkg.dev/nais-io/nais/images/myapp:1.0.0",
			ingresses:      []string{"myapp.prod-gcp.nais.io"},
			expectedState:  model.StateNotnais,
			expectedErrors: []model.ErrorType{model.ErrorTypeDeprecatedIngress},
		},
	}

	for _, tc := range testCases {
		app := &model.App{Image: tc.image, Ingresses: tc.ingresses, Env: &model.Env{Name: "prod-gcp"}}
		fmt.Println(tc.name)
		setStatus(app, []metav1.Condition{{Status: metav1.ConditionTrue, Reason: string(tc.appCondition)}}, asInstances(tc.instanceStates))
		if app.AppState.State != tc.expectedState {
			t.Errorf("%s\ngot state: %v, want: %v", tc.name, app.AppState.State, tc.expectedState)
		}
		if !hasError(app.AppState.Errors, tc.expectedErrors) {
			t.Errorf("%s\ngot error: %v, want: %v", tc.name, &app.AppState.Errors, tc.expectedErrors)
		}
	}
}

func hasError(errors []*model.StateError, expectedErrors []model.ErrorType) bool {
	if len(errors) != len(expectedErrors) {
		return false
	}

	for _, error := range expectedErrors {
		if !contains(errors, error) {
			return false
		}
	}
	return true
}

func contains(slice []*model.StateError, s model.ErrorType) bool {
	for _, item := range slice {
		if item.Type == s {
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
