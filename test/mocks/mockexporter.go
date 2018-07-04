// Code generated by MockGen. DO NOT EDIT.
// Source: go.opencensus.io/stats/view (interfaces: Exporter)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	view "go.opencensus.io/stats/view"
)

// MockExporter is a mock of Exporter interface
type MockExporter struct {
	ctrl     *gomock.Controller
	recorder *MockExporterMockRecorder
}

// MockExporterMockRecorder is the mock recorder for MockExporter
type MockExporterMockRecorder struct {
	mock *MockExporter
}

// NewMockExporter creates a new mock instance
func NewMockExporter(ctrl *gomock.Controller) *MockExporter {
	mock := &MockExporter{ctrl: ctrl}
	mock.recorder = &MockExporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExporter) EXPECT() *MockExporterMockRecorder {
	return m.recorder
}

// ExportView mocks base method
func (m *MockExporter) ExportView(arg0 *view.Data) {
	m.ctrl.Call(m, "ExportView", arg0)
}

// ExportView indicates an expected call of ExportView
func (mr *MockExporterMockRecorder) ExportView(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExportView", reflect.TypeOf((*MockExporter)(nil).ExportView), arg0)
}