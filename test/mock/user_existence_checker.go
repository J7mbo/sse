// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import internal "sse/internal"
import mock "github.com/stretchr/testify/mock"

// userExistenceChecker is an autogenerated mock type for the userExistenceChecker type
type UserExistenceChecker struct {
	mock.Mock
}

// UserExistsInDB provides a mock function with given fields: ui
func (_m *UserExistenceChecker) UserExistsInDB(ui *internal.UserInfo) (bool, error) {
	ret := _m.Called(ui)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*internal.UserInfo) bool); ok {
		r0 = rf(ui)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*internal.UserInfo) error); ok {
		r1 = rf(ui)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}