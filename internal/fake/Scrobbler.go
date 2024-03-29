// Code generated by mockery v2.40.1. DO NOT EDIT.

package fake

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	pianobar "github.com/nlowe/pianoman/pianobar"
)

// Scrobbler is an autogenerated mock type for the Scrobbler type
type Scrobbler struct {
	mock.Mock
}

type Scrobbler_Expecter struct {
	mock *mock.Mock
}

func (_m *Scrobbler) EXPECT() *Scrobbler_Expecter {
	return &Scrobbler_Expecter{mock: &_m.Mock}
}

// Scrobble provides a mock function with given fields: ctx, t
func (_m *Scrobbler) Scrobble(ctx context.Context, t ...pianobar.Track) error {
	_va := make([]interface{}, len(t))
	for _i := range t {
		_va[_i] = t[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Scrobble")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...pianobar.Track) error); ok {
		r0 = rf(ctx, t...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Scrobbler_Scrobble_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Scrobble'
type Scrobbler_Scrobble_Call struct {
	*mock.Call
}

// Scrobble is a helper method to define mock.On call
//   - ctx context.Context
//   - t ...pianobar.Track
func (_e *Scrobbler_Expecter) Scrobble(ctx interface{}, t ...interface{}) *Scrobbler_Scrobble_Call {
	return &Scrobbler_Scrobble_Call{Call: _e.mock.On("Scrobble",
		append([]interface{}{ctx}, t...)...)}
}

func (_c *Scrobbler_Scrobble_Call) Run(run func(ctx context.Context, t ...pianobar.Track)) *Scrobbler_Scrobble_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]pianobar.Track, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(pianobar.Track)
			}
		}
		run(args[0].(context.Context), variadicArgs...)
	})
	return _c
}

func (_c *Scrobbler_Scrobble_Call) Return(_a0 error) *Scrobbler_Scrobble_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Scrobbler_Scrobble_Call) RunAndReturn(run func(context.Context, ...pianobar.Track) error) *Scrobbler_Scrobble_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateNowPlaying provides a mock function with given fields: ctx, t
func (_m *Scrobbler) UpdateNowPlaying(ctx context.Context, t pianobar.Track) error {
	ret := _m.Called(ctx, t)

	if len(ret) == 0 {
		panic("no return value specified for UpdateNowPlaying")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, pianobar.Track) error); ok {
		r0 = rf(ctx, t)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Scrobbler_UpdateNowPlaying_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateNowPlaying'
type Scrobbler_UpdateNowPlaying_Call struct {
	*mock.Call
}

// UpdateNowPlaying is a helper method to define mock.On call
//   - ctx context.Context
//   - t pianobar.Track
func (_e *Scrobbler_Expecter) UpdateNowPlaying(ctx interface{}, t interface{}) *Scrobbler_UpdateNowPlaying_Call {
	return &Scrobbler_UpdateNowPlaying_Call{Call: _e.mock.On("UpdateNowPlaying", ctx, t)}
}

func (_c *Scrobbler_UpdateNowPlaying_Call) Run(run func(ctx context.Context, t pianobar.Track)) *Scrobbler_UpdateNowPlaying_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pianobar.Track))
	})
	return _c
}

func (_c *Scrobbler_UpdateNowPlaying_Call) Return(_a0 error) *Scrobbler_UpdateNowPlaying_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Scrobbler_UpdateNowPlaying_Call) RunAndReturn(run func(context.Context, pianobar.Track) error) *Scrobbler_UpdateNowPlaying_Call {
	_c.Call.Return(run)
	return _c
}

// NewScrobbler creates a new instance of Scrobbler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewScrobbler(t interface {
	mock.TestingT
	Cleanup(func())
}) *Scrobbler {
	mock := &Scrobbler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
