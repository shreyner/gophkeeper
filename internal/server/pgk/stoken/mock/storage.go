// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/server/pgk/stoken/interface.go

// Package stoken is a generated GoMock package.
package stoken

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	stoken "github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
)

// MockJWTService is a mock of JWTService interface.
type MockJWTService struct {
	ctrl     *gomock.Controller
	recorder *MockJWTServiceMockRecorder
}

// MockJWTServiceMockRecorder is the mock recorder for MockJWTService.
type MockJWTServiceMockRecorder struct {
	mock *MockJWTService
}

// NewMockJWTService creates a new mock instance.
func NewMockJWTService(ctrl *gomock.Controller) *MockJWTService {
	mock := &MockJWTService{ctrl: ctrl}
	mock.recorder = &MockJWTServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJWTService) EXPECT() *MockJWTServiceMockRecorder {
	return m.recorder
}

// CreateToken mocks base method.
func (m *MockJWTService) CreateToken(data *stoken.Data) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateToken", data)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateToken indicates an expected call of CreateToken.
func (mr *MockJWTServiceMockRecorder) CreateToken(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateToken", reflect.TypeOf((*MockJWTService)(nil).CreateToken), data)
}

// ParseToken mocks base method.
func (m *MockJWTService) ParseToken(tokenString string) (*stoken.Data, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseToken", tokenString)
	ret0, _ := ret[0].(*stoken.Data)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseToken indicates an expected call of ParseToken.
func (mr *MockJWTServiceMockRecorder) ParseToken(tokenString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseToken", reflect.TypeOf((*MockJWTService)(nil).ParseToken), tokenString)
}