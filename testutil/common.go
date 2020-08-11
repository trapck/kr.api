package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//FailOnEqual stops test if condition is matched
func FailOnEqual(t *testing.T, v1 interface{}, v2 interface{}, msg string) {
	t.Helper()
	if v1 == v2 {
		assert.FailNow(t, msg)
	}
}

//FailOnNotEqual stops test if condition is not matched
func FailOnNotEqual(t *testing.T, v1 interface{}, v2 interface{}, msg string) {
	t.Helper()
	if v1 != v2 {
		assert.FailNow(t, msg)
	}
}
