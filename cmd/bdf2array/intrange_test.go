package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntRangeSuccess(t *testing.T) {
	type testSet struct {
		name           string
		arguments      []string
		expectedValues []int
		expectedString string
	}

	testSets := []testSet{
		{"empty", []string{}, []int{}, ""},
		{"1", []string{"1"}, []int{1}, "0x01"},
		{"0x01", []string{"0x01"}, []int{1}, "0x01"},
		{"twice the same", []string{"1", "1"}, []int{1}, "0x01"},
		{"1,2", []string{"1,2"}, []int{1, 2}, "0x01,0x02"},
		{"1,2 separate", []string{"1", "2"}, []int{1, 2}, "0x01,0x02"},
		{"1-2", []string{"1-2"}, []int{1, 2}, "0x01,0x02"},
		{"1-2", []string{"1-2"}, []int{1, 2}, "0x01,0x02"},
		{"1,2,3", []string{"1,2,3"}, []int{1, 2, 3}, "0x01-0x03"},
		{"1-3", []string{"1-3"}, []int{1, 2, 3}, "0x01-0x03"},
		{"0x01-0x03", []string{"0x01-0x03"}, []int{1, 2, 3}, "0x01-0x03"},
		{"1,2,3 separate", []string{"1", "2", "3"}, []int{1, 2, 3}, "0x01-0x03"},
		{"1,2,3,5", []string{"1-3,5"}, []int{1, 2, 3, 5}, "0x01-0x03,0x05"},
		{"1,2,3,5 separate", []string{"1-3", "5"}, []int{1, 2, 3, 5}, "0x01-0x03,0x05"},
		{"1,2,3,5,7", []string{"1-3,5,7"}, []int{1, 2, 3, 5, 7}, "0x01-0x03,0x05,0x07"},
		{"1,2,3,5,6,8", []string{"1-3,5,6,8"}, []int{1, 2, 3, 5, 6, 8}, "0x01-0x03,0x05,0x06,0x08"},
	}

	for _, test := range testSets {
		test := test // keep 'test' in scope for the closure function
		t.Run(test.name, func(t *testing.T) {
			underTest := make(intrange)

			// set values:
			for _, arg := range test.arguments {
				assert.NoError(t, underTest.Set(arg))
			}

			// check results:
			assert.Equal(t, test.expectedValues, underTest.GetSlice())
			assert.Equal(t, test.expectedString, underTest.String())
		})
	}
}

func TestIntRangeNilReceiversSet(t *testing.T) {
	// var exists, map was not created:
	var underTest1 intrange
	assert.NoError(t, underTest1.Set("1"))

	// nil pointer:
	var underTest2 *intrange
	err := underTest2.Set("1")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "nil receiver")
	}
}

func TestIntRangeNilReceiversGetSlice(t *testing.T) {
	// var exists, map was not created:
	var underTest1 intrange
	assert.Equal(t, []int{}, underTest1.GetSlice())

	// nil pointer:
	var underTest2 *intrange
	assert.Equal(t, []int{}, underTest2.GetSlice())
}

func TestIntRangeNilReceiversString(t *testing.T) {
	// var exists, map was not created:
	var underTest1 intrange
	assert.Equal(t, "nil", underTest1.String())

	// nil pointer:
	var underTest2 *intrange
	assert.Equal(t, "nil", underTest2.String())
}

func TestIntRangeSetErrors(t *testing.T) {
	type testSet struct {
		name          string
		arguments     []string
		expectedError string
	}

	testSets := []testSet{
		{"empty arg", []string{""}, "invalid syntax"},
		{"not a number", []string{"a"}, "invalid syntax"},
		{"second arg not a number", []string{"1-a"}, "invalid syntax"},
		{"inverted range", []string{"3-1"}, "invalid range specified"},
		{"double dash", []string{"1-3-5"}, "contains too many '-'"},
	}

	for _, test := range testSets {
		test := test // keep 'test' in scope for the closure function
		t.Run(test.name, func(t *testing.T) {
			underTest := make(intrange)

			// set values:
			for _, arg := range test.arguments {
				err := underTest.Set(arg)
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), test.expectedError)
				}
			}
		})
	}
}
