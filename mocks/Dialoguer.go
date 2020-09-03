// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	slack "github.com/slack-go/slack"
	mock "github.com/stretchr/testify/mock"
)

// Dialoguer is an autogenerated mock type for the Dialoguer type
type Dialoguer struct {
	mock.Mock
}

// OpenView provides a mock function with given fields: triggerID, view
func (_m *Dialoguer) OpenView(triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error) {
	ret := _m.Called(triggerID, view)

	var r0 *slack.ViewResponse
	if rf, ok := ret.Get(0).(func(string, slack.ModalViewRequest) *slack.ViewResponse); ok {
		r0 = rf(triggerID, view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*slack.ViewResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, slack.ModalViewRequest) error); ok {
		r1 = rf(triggerID, view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}