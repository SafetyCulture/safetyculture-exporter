// Code generated by mockery v2.45.1. DO NOT EDIT.

package mocks

import (
	feed "github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Exporter is an autogenerated mock type for the Exporter type
type Exporter struct {
	mock.Mock
}

// CreateSchema provides a mock function with given fields: _a0, rows
func (_m *Exporter) CreateSchema(_a0 feed.Feed, rows interface{}) error {
	ret := _m.Called(_a0, rows)

	if len(ret) == 0 {
		panic("no return value specified for CreateSchema")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(feed.Feed, interface{}) error); ok {
		r0 = rf(_a0, rows)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteRowsIfExist provides a mock function with given fields: _a0, query, args
func (_m *Exporter) DeleteRowsIfExist(_a0 feed.Feed, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, _a0, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRowsIfExist")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(feed.Feed, string, ...interface{}) error); ok {
		r0 = rf(_a0, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FinaliseExport provides a mock function with given fields: _a0, rows
func (_m *Exporter) FinaliseExport(_a0 feed.Feed, rows interface{}) error {
	ret := _m.Called(_a0, rows)

	if len(ret) == 0 {
		panic("no return value specified for FinaliseExport")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(feed.Feed, interface{}) error); ok {
		r0 = rf(_a0, rows)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDuration provides a mock function with given fields:
func (_m *Exporter) GetDuration() time.Duration {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetDuration")
	}

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// InitFeed provides a mock function with given fields: _a0, opts
func (_m *Exporter) InitFeed(_a0 feed.Feed, opts *feed.InitFeedOptions) error {
	ret := _m.Called(_a0, opts)

	if len(ret) == 0 {
		panic("no return value specified for InitFeed")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(feed.Feed, *feed.InitFeedOptions) error); ok {
		r0 = rf(_a0, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LastModifiedAt provides a mock function with given fields: _a0, modifiedAfter, columnName, orgID
func (_m *Exporter) LastModifiedAt(_a0 feed.Feed, modifiedAfter time.Time, columnName string, orgID string) (time.Time, error) {
	ret := _m.Called(_a0, modifiedAfter, columnName, orgID)

	if len(ret) == 0 {
		panic("no return value specified for LastModifiedAt")
	}

	var r0 time.Time
	var r1 error
	if rf, ok := ret.Get(0).(func(feed.Feed, time.Time, string, string) (time.Time, error)); ok {
		return rf(_a0, modifiedAfter, columnName, orgID)
	}
	if rf, ok := ret.Get(0).(func(feed.Feed, time.Time, string, string) time.Time); ok {
		r0 = rf(_a0, modifiedAfter, columnName, orgID)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	if rf, ok := ret.Get(1).(func(feed.Feed, time.Time, string, string) error); ok {
		r1 = rf(_a0, modifiedAfter, columnName, orgID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LastRecord provides a mock function with given fields: _a0, modifiedAfter, orgID, sortColumn
func (_m *Exporter) LastRecord(_a0 feed.Feed, modifiedAfter time.Time, orgID string, sortColumn string) time.Time {
	ret := _m.Called(_a0, modifiedAfter, orgID, sortColumn)

	if len(ret) == 0 {
		panic("no return value specified for LastRecord")
	}

	var r0 time.Time
	if rf, ok := ret.Get(0).(func(feed.Feed, time.Time, string, string) time.Time); ok {
		r0 = rf(_a0, modifiedAfter, orgID, sortColumn)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// ParameterLimit provides a mock function with given fields:
func (_m *Exporter) ParameterLimit() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ParameterLimit")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// SupportsUpsert provides a mock function with given fields:
func (_m *Exporter) SupportsUpsert() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for SupportsUpsert")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// UpdateRows provides a mock function with given fields: _a0, primaryKeys, element
func (_m *Exporter) UpdateRows(_a0 feed.Feed, primaryKeys []string, element map[string]interface{}) (int64, error) {
	ret := _m.Called(_a0, primaryKeys, element)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRows")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(feed.Feed, []string, map[string]interface{}) (int64, error)); ok {
		return rf(_a0, primaryKeys, element)
	}
	if rf, ok := ret.Get(0).(func(feed.Feed, []string, map[string]interface{}) int64); ok {
		r0 = rf(_a0, primaryKeys, element)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(feed.Feed, []string, map[string]interface{}) error); ok {
		r1 = rf(_a0, primaryKeys, element)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteMedia provides a mock function with given fields: auditID, mediaID, contentType, body
func (_m *Exporter) WriteMedia(auditID string, mediaID string, contentType string, body []byte) error {
	ret := _m.Called(auditID, mediaID, contentType, body)

	if len(ret) == 0 {
		panic("no return value specified for WriteMedia")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, []byte) error); ok {
		r0 = rf(auditID, mediaID, contentType, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteRows provides a mock function with given fields: _a0, rows
func (_m *Exporter) WriteRows(_a0 feed.Feed, rows interface{}) error {
	ret := _m.Called(_a0, rows)

	if len(ret) == 0 {
		panic("no return value specified for WriteRows")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(feed.Feed, interface{}) error); ok {
		r0 = rf(_a0, rows)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewExporter creates a new instance of Exporter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExporter(t interface {
	mock.TestingT
	Cleanup(func())
}) *Exporter {
	mock := &Exporter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
