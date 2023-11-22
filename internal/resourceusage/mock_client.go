// Code generated by mockery. DO NOT EDIT.

package resourceusage

import (
	context "context"

	model "github.com/nais/console-backend/internal/graph/model"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// CurrentResourceUtilizationForApp provides a mock function with given fields: ctx, env, team, app
func (_m *MockClient) CurrentResourceUtilizationForApp(ctx context.Context, env string, team string, app string) (*model.CurrentResourceUtilizationForApp, error) {
	ret := _m.Called(ctx, env, team, app)

	var r0 *model.CurrentResourceUtilizationForApp
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*model.CurrentResourceUtilizationForApp, error)); ok {
		return rf(ctx, env, team, app)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.CurrentResourceUtilizationForApp); ok {
		r0 = rf(ctx, env, team, app)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.CurrentResourceUtilizationForApp)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, env, team, app)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_CurrentResourceUtilizationForApp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CurrentResourceUtilizationForApp'
type MockClient_CurrentResourceUtilizationForApp_Call struct {
	*mock.Call
}

// CurrentResourceUtilizationForApp is a helper method to define mock.On call
//   - ctx context.Context
//   - env string
//   - team string
//   - app string
func (_e *MockClient_Expecter) CurrentResourceUtilizationForApp(ctx interface{}, env interface{}, team interface{}, app interface{}) *MockClient_CurrentResourceUtilizationForApp_Call {
	return &MockClient_CurrentResourceUtilizationForApp_Call{Call: _e.mock.On("CurrentResourceUtilizationForApp", ctx, env, team, app)}
}

func (_c *MockClient_CurrentResourceUtilizationForApp_Call) Run(run func(ctx context.Context, env string, team string, app string)) *MockClient_CurrentResourceUtilizationForApp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *MockClient_CurrentResourceUtilizationForApp_Call) Return(_a0 *model.CurrentResourceUtilizationForApp, _a1 error) *MockClient_CurrentResourceUtilizationForApp_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_CurrentResourceUtilizationForApp_Call) RunAndReturn(run func(context.Context, string, string, string) (*model.CurrentResourceUtilizationForApp, error)) *MockClient_CurrentResourceUtilizationForApp_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceUtilizationForApp provides a mock function with given fields: ctx, env, team, app, start, end
func (_m *MockClient) ResourceUtilizationForApp(ctx context.Context, env string, team string, app string, start time.Time, end time.Time) (*model.ResourceUtilizationForApp, error) {
	ret := _m.Called(ctx, env, team, app, start, end)

	var r0 *model.ResourceUtilizationForApp
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Time, time.Time) (*model.ResourceUtilizationForApp, error)); ok {
		return rf(ctx, env, team, app, start, end)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Time, time.Time) *model.ResourceUtilizationForApp); ok {
		r0 = rf(ctx, env, team, app, start, end)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ResourceUtilizationForApp)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, time.Time, time.Time) error); ok {
		r1 = rf(ctx, env, team, app, start, end)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ResourceUtilizationForApp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceUtilizationForApp'
type MockClient_ResourceUtilizationForApp_Call struct {
	*mock.Call
}

// ResourceUtilizationForApp is a helper method to define mock.On call
//   - ctx context.Context
//   - env string
//   - team string
//   - app string
//   - start time.Time
//   - end time.Time
func (_e *MockClient_Expecter) ResourceUtilizationForApp(ctx interface{}, env interface{}, team interface{}, app interface{}, start interface{}, end interface{}) *MockClient_ResourceUtilizationForApp_Call {
	return &MockClient_ResourceUtilizationForApp_Call{Call: _e.mock.On("ResourceUtilizationForApp", ctx, env, team, app, start, end)}
}

func (_c *MockClient_ResourceUtilizationForApp_Call) Run(run func(ctx context.Context, env string, team string, app string, start time.Time, end time.Time)) *MockClient_ResourceUtilizationForApp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string), args[4].(time.Time), args[5].(time.Time))
	})
	return _c
}

func (_c *MockClient_ResourceUtilizationForApp_Call) Return(_a0 *model.ResourceUtilizationForApp, _a1 error) *MockClient_ResourceUtilizationForApp_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ResourceUtilizationForApp_Call) RunAndReturn(run func(context.Context, string, string, string, time.Time, time.Time) (*model.ResourceUtilizationForApp, error)) *MockClient_ResourceUtilizationForApp_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceUtilizationForTeam provides a mock function with given fields: ctx, team, start, end
func (_m *MockClient) ResourceUtilizationForTeam(ctx context.Context, team string, start time.Time, end time.Time) ([]model.ResourceUtilizationForEnv, error) {
	ret := _m.Called(ctx, team, start, end)

	var r0 []model.ResourceUtilizationForEnv
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) ([]model.ResourceUtilizationForEnv, error)); ok {
		return rf(ctx, team, start, end)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) []model.ResourceUtilizationForEnv); ok {
		r0 = rf(ctx, team, start, end)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ResourceUtilizationForEnv)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Time, time.Time) error); ok {
		r1 = rf(ctx, team, start, end)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ResourceUtilizationForTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceUtilizationForTeam'
type MockClient_ResourceUtilizationForTeam_Call struct {
	*mock.Call
}

// ResourceUtilizationForTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - team string
//   - start time.Time
//   - end time.Time
func (_e *MockClient_Expecter) ResourceUtilizationForTeam(ctx interface{}, team interface{}, start interface{}, end interface{}) *MockClient_ResourceUtilizationForTeam_Call {
	return &MockClient_ResourceUtilizationForTeam_Call{Call: _e.mock.On("ResourceUtilizationForTeam", ctx, team, start, end)}
}

func (_c *MockClient_ResourceUtilizationForTeam_Call) Run(run func(ctx context.Context, team string, start time.Time, end time.Time)) *MockClient_ResourceUtilizationForTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(time.Time), args[3].(time.Time))
	})
	return _c
}

func (_c *MockClient_ResourceUtilizationForTeam_Call) Return(_a0 []model.ResourceUtilizationForEnv, _a1 error) *MockClient_ResourceUtilizationForTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ResourceUtilizationForTeam_Call) RunAndReturn(run func(context.Context, string, time.Time, time.Time) ([]model.ResourceUtilizationForEnv, error)) *MockClient_ResourceUtilizationForTeam_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceUtilizationOverageForTeam provides a mock function with given fields: ctx, team, start, end
func (_m *MockClient) ResourceUtilizationOverageForTeam(ctx context.Context, team string, start time.Time, end time.Time) (*model.ResourceUtilizationOverageForTeam, error) {
	ret := _m.Called(ctx, team, start, end)

	var r0 *model.ResourceUtilizationOverageForTeam
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) (*model.ResourceUtilizationOverageForTeam, error)); ok {
		return rf(ctx, team, start, end)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, time.Time) *model.ResourceUtilizationOverageForTeam); ok {
		r0 = rf(ctx, team, start, end)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ResourceUtilizationOverageForTeam)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Time, time.Time) error); ok {
		r1 = rf(ctx, team, start, end)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ResourceUtilizationOverageForTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceUtilizationOverageForTeam'
type MockClient_ResourceUtilizationOverageForTeam_Call struct {
	*mock.Call
}

// ResourceUtilizationOverageForTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - team string
//   - start time.Time
//   - end time.Time
func (_e *MockClient_Expecter) ResourceUtilizationOverageForTeam(ctx interface{}, team interface{}, start interface{}, end interface{}) *MockClient_ResourceUtilizationOverageForTeam_Call {
	return &MockClient_ResourceUtilizationOverageForTeam_Call{Call: _e.mock.On("ResourceUtilizationOverageForTeam", ctx, team, start, end)}
}

func (_c *MockClient_ResourceUtilizationOverageForTeam_Call) Run(run func(ctx context.Context, team string, start time.Time, end time.Time)) *MockClient_ResourceUtilizationOverageForTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(time.Time), args[3].(time.Time))
	})
	return _c
}

func (_c *MockClient_ResourceUtilizationOverageForTeam_Call) Return(_a0 *model.ResourceUtilizationOverageForTeam, _a1 error) *MockClient_ResourceUtilizationOverageForTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ResourceUtilizationOverageForTeam_Call) RunAndReturn(run func(context.Context, string, time.Time, time.Time) (*model.ResourceUtilizationOverageForTeam, error)) *MockClient_ResourceUtilizationOverageForTeam_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceUtilizationRangeForApp provides a mock function with given fields: ctx, env, team, app
func (_m *MockClient) ResourceUtilizationRangeForApp(ctx context.Context, env string, team string, app string) (*model.ResourceUtilizationDateRange, error) {
	ret := _m.Called(ctx, env, team, app)

	var r0 *model.ResourceUtilizationDateRange
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*model.ResourceUtilizationDateRange, error)); ok {
		return rf(ctx, env, team, app)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.ResourceUtilizationDateRange); ok {
		r0 = rf(ctx, env, team, app)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ResourceUtilizationDateRange)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, env, team, app)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ResourceUtilizationRangeForApp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceUtilizationRangeForApp'
type MockClient_ResourceUtilizationRangeForApp_Call struct {
	*mock.Call
}

// ResourceUtilizationRangeForApp is a helper method to define mock.On call
//   - ctx context.Context
//   - env string
//   - team string
//   - app string
func (_e *MockClient_Expecter) ResourceUtilizationRangeForApp(ctx interface{}, env interface{}, team interface{}, app interface{}) *MockClient_ResourceUtilizationRangeForApp_Call {
	return &MockClient_ResourceUtilizationRangeForApp_Call{Call: _e.mock.On("ResourceUtilizationRangeForApp", ctx, env, team, app)}
}

func (_c *MockClient_ResourceUtilizationRangeForApp_Call) Run(run func(ctx context.Context, env string, team string, app string)) *MockClient_ResourceUtilizationRangeForApp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *MockClient_ResourceUtilizationRangeForApp_Call) Return(_a0 *model.ResourceUtilizationDateRange, _a1 error) *MockClient_ResourceUtilizationRangeForApp_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ResourceUtilizationRangeForApp_Call) RunAndReturn(run func(context.Context, string, string, string) (*model.ResourceUtilizationDateRange, error)) *MockClient_ResourceUtilizationRangeForApp_Call {
	_c.Call.Return(run)
	return _c
}

// ResourceUtilizationRangeForTeam provides a mock function with given fields: ctx, team
func (_m *MockClient) ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error) {
	ret := _m.Called(ctx, team)

	var r0 *model.ResourceUtilizationDateRange
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.ResourceUtilizationDateRange, error)); ok {
		return rf(ctx, team)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.ResourceUtilizationDateRange); ok {
		r0 = rf(ctx, team)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ResourceUtilizationDateRange)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, team)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ResourceUtilizationRangeForTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResourceUtilizationRangeForTeam'
type MockClient_ResourceUtilizationRangeForTeam_Call struct {
	*mock.Call
}

// ResourceUtilizationRangeForTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - team string
func (_e *MockClient_Expecter) ResourceUtilizationRangeForTeam(ctx interface{}, team interface{}) *MockClient_ResourceUtilizationRangeForTeam_Call {
	return &MockClient_ResourceUtilizationRangeForTeam_Call{Call: _e.mock.On("ResourceUtilizationRangeForTeam", ctx, team)}
}

func (_c *MockClient_ResourceUtilizationRangeForTeam_Call) Run(run func(ctx context.Context, team string)) *MockClient_ResourceUtilizationRangeForTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockClient_ResourceUtilizationRangeForTeam_Call) Return(_a0 *model.ResourceUtilizationDateRange, _a1 error) *MockClient_ResourceUtilizationRangeForTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ResourceUtilizationRangeForTeam_Call) RunAndReturn(run func(context.Context, string) (*model.ResourceUtilizationDateRange, error)) *MockClient_ResourceUtilizationRangeForTeam_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
