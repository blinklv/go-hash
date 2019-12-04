// hash_test.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-12-04
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-12-04

package main

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestCRun(t *testing.T) {
	for _, env := range []struct {
		cnum    int
		counter int64
		result  int64
	}{
		{-1, 0, 0},
		{0, 0, 0},
		{1, 0, 100},
		{16, 0, 1600},
		{32, 0, 3200},
	} {
		crun(env.cnum, func() {
			for i := 0; i < 100; i++ {
				atomic.AddInt64(&(env.counter), 1)
			}
		})
		a := &assertions{assert.New(t), env}
		a.equalf(env.result, env.counter, "counter:%d != result:%d", env.counter, env.result)
	}
}

// Wraps assert.Assertions to output environment information.
type assertions struct {
	*assert.Assertions
	env interface{}
}

func (a *assertions) equalf(expected, actual interface{}, msg string, args ...interface{}) bool {
	return a.Equalf(int2int64(expected), int2int64(actual), sprintf("%#v %s", a.env, msg), args...)
}

func (a *assertions) not_equalf(expected, actual interface{}, msg string, args ...interface{}) bool {
	return a.NotEqualf(int2int64(expected), int2int64(actual), sprintf("%#v %s", a.env, msg), args...)
}

func int2int64(i interface{}) interface{} {
	if v, ok := i.(int); ok {
		return int64(v)
	}
	return i
}
