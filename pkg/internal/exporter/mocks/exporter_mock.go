// Code generated by mockery v2.4.0. DO NOT EDIT.

package exportermock

import (
	json "encoding/json"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Exporter is an autogenerated mock type for the Exporter type
type Exporter struct {
	mock.Mock
}

// GetLastModifiedAt provides a mock function with given fields: modifiedAfter
func (_m *Exporter) GetLastModifiedAt(modifiedAfter time.Time) *time.Time {
	ret := _m.Called(modifiedAfter)

	var r0 *time.Time
	if rf, ok := ret.Get(0).(func(time.Time) *time.Time); ok {
		r0 = rf(modifiedAfter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*time.Time)
		}
	}

	return r0
}

// SetLastModifiedAt provides a mock function with given fields: modifiedAt
func (_m *Exporter) SetLastModifiedAt(modifiedAt time.Time) {
	_m.Called(modifiedAt)
}

// WriteRow provides a mock function with given fields: name, row
func (_m *Exporter) WriteRow(name string, row *json.RawMessage) {
	_m.Called(name, row)
}
