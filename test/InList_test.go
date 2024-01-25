package test

import (
	"jinx/pkg/util"
	"strings"
	"testing"
)

func TestInListWithSlice(t *testing.T) {

	tests := []struct {
		list      []string
		element   string
		predicate func(a string, b string) bool
		dsp       string
		expect    bool
	}{
		{
			[]string{"Martin", "Kevin", "Myriam"},
			"Martin",
			func(a string, b string) bool {
				return a == b
			},
			"should return true",
			true,
		},

		{
			[]string{"Martin", "Kevin", "Myriam"},
			"martin",
			func(a string, b string) bool {
				return strings.ToLower(a) == strings.ToLower(b)
			},
			"should return true",
			true,
		},

		{
			[]string{"Martin", "Kevin", "Myriam"},
			"Lydia",
			func(a string, b string) bool {
				return strings.ToLower(a) == strings.ToLower(b)
			},
			"should return false",
			false,
		},

		{
			[]string{},
			"Lydia",
			func(a string, b string) bool {
				return strings.ToLower(a) == strings.ToLower(b)
			},
			"should return false",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.dsp, func(t *testing.T) {
			result := util.InList[string](test.list, test.element, test.predicate)
			if result != test.expect {
				t.Errorf("expected %v got %v", test.expect, result)
			}
		})
	}
}

func TestInListWithStruct(t *testing.T) {
	type S struct {
		name string
		age  int
	}
	tests := []struct {
		list      []S
		element   S
		predicate func(a S, b S) bool
		dsp       string
		expect    bool
	}{
		{
			[]S{
				{"Martin", 25},
				{"Collins", 23},
				{"Myriam", 21},
				{"Marcel", 19},
			},
			S{name: "Martin", age: 25},
			func(a S, b S) bool {
				return a.name == b.name && a.age == b.age
			},
			"should return true",
			true,
		},

		{
			[]S{
				{"Martin", 25},
				{"Collins", 23},
				{"Myriam", 21},
				{"Marcel", 19},
			},
			S{name: "Collins", age: 23},
			func(a S, b S) bool {
				return a.name == b.name && a.age == b.age
			},
			"should return true",
			true,
		},

		{
			[]S{
				{"Martin", 25},
				{"Collins", 23},
				{"Myriam", 21},
				{"Marcel", 19},
			},
			S{name: "Lydia", age: 25},
			func(a S, b S) bool {
				return a.name == b.name && a.age == b.age
			},
			"should return false",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.dsp, func(t *testing.T) {
			result := util.InList[S](test.list, test.element, test.predicate)
			if result != test.expect {
				t.Errorf("expected %v got %v", test.expect, result)
			}
		})
	}

}
