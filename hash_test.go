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

func TestPad(t *testing.T) {
	for _, env := range []struct {
		str    string
		width  int
		result string
	}{
		{"", -1, ""},
		{"Hello", -1, "Hello"},
		{"", 0, ""},
		{"Hello", 0, "Hello"},
		{"", 1, " "},
		{"Hello", 1, "Hello"},
		{"", 10, "          "},
		{"Hello", 10, "Hello     "},
	} {
		result := pad(env.str, env.width)
		a := assert.New(t)
		a.Equalf(env.result, result, "%#v", env)
	}
}

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
		a := assert.New(t)
		a.Equalf(env.result, env.counter, "%#v", env)
	}
}
