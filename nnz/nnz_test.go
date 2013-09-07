package nnz

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
)

type jsonAndSQLSerializable interface {
	driver.Valuer
	sql.Scanner
	json.Marshaler
	json.Unmarshaler

	deref() interface{}
}

func intPtr(i Int) *Int {
	return &i
}

func (i *Int) deref() interface{} {
	return *i
}

func stringPtr(s String) *String {
	return &s
}

func (s *String) deref() interface{} {
	return *s
}

func TestNNZTypes(t *testing.T) {
	tests := []struct {
		value    jsonAndSQLSerializable
		empty    jsonAndSQLSerializable
		jsonRepr string
		sqlValue driver.Value
	}{
		{intPtr(123), new(Int), "123", int(123)},
		{intPtr(0), new(Int), "null", nil},
		{stringPtr("abc"), new(String), `"abc"`, "abc"},
		{stringPtr(""), new(String), "null", nil},
	}

	for _, test := range tests {
		// MarshalJSON
		jsonReprBytes, err := json.Marshal(test.value)
		if err != nil {
			t.Errorf("json.Marshal(%v): %s", test.value, err)
			continue
		}
		jsonRepr := string(jsonReprBytes)
		if test.jsonRepr != jsonRepr {
			t.Errorf("%v: want jsonRepr == %q, got %q", test.value, test.jsonRepr, jsonRepr)
		}

		// UnmarshalJSON
		var valueFromJSON = test.empty
		err = json.Unmarshal(jsonReprBytes, valueFromJSON)
		if err != nil {
			t.Errorf("json.Unmarshal(%s, _): %s", jsonReprBytes, err)
			continue
		}
		if !reflect.DeepEqual(test.value.deref(), valueFromJSON.deref()) {
			t.Errorf("%v: want valueFromJSON == %v, got %v", test.value.deref(), test.value.deref(), valueFromJSON.deref())
		}

		// driver.Value
		sqlValue, err := test.value.Value()
		if err != nil {
			t.Errorf("(%v).Value(): %s", test.value, err)
			continue
		}
		if !reflect.DeepEqual(test.sqlValue, sqlValue) {
			t.Errorf("%v: want sqlValue == %v (%T), got %v (%T)", test.value, test.sqlValue, test.sqlValue, sqlValue, sqlValue)
		}

		// driver.Scan
		var valueFromSQL = test.empty
		err = valueFromSQL.Scan(sqlValue)
		if err != nil {
			t.Errorf("(%T).Scan(%v): %s", test.value, sqlValue, err)
			continue
		}
		if !reflect.DeepEqual(test.value.deref(), valueFromSQL.deref()) {
			t.Errorf("%v: want valueFromSQL == %v, got %v", test.value.deref(), test.value.deref(), valueFromSQL.deref())
		}
	}
}
