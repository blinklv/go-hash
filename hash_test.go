// hash_test.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-12-04
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-12-05

package main

import (
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestRsort(t *testing.T) {
	for _, env := range []struct {
		strs   []string
		result []string
	}{
		{nil, nil},
		{[]string{}, []string{}},
		{[]string{"a"}, []string{"a"}},
		{
			[]string{"a", "b", "foo", "car", "hello", "world"},
			[]string{"world", "hello", "foo", "car", "b", "a"},
		},
	} {
		rsort(env.strs)
		a := assert.New(t)
		a.Equalf(env.result, env.strs, "%+v", env)
	}
}

func TestSplit(t *testing.T) {
	for _, env := range []struct {
		rest  string
		width int
		lines []string
	}{
		{"", -1, []string{}},
		{"Hello", -1, []string{}},
		{"", 0, []string{}},
		{"Hello", 0, []string{}},
		{"", 1, []string{}},
		{"Hello", 1, []string{"H", "e", "l", "l", "o"}},
		{"What do you want to do?", 4, []string{"What", " do ", "you ", "want", " to ", "do?"}},
	} {
		lines := split(env.rest, env.width)
		a := assert.New(t)
		a.Equalf(env.lines, lines, "%+v", env)
	}
}

func TestCut(t *testing.T) {
	for _, env := range []struct {
		str  string
		cp   int
		a, b string
	}{
		{"", -1, "", ""},
		{"", 0, "", ""},
		{"", 1, "", ""},
		{"Hello", -1, "", "Hello"},
		{"Hello", 0, "", "Hello"},
		{"Hello", 1, "H", "ello"},
		{"Hello", 3, "Hel", "lo"},
		{"Hello", 5, "Hello", ""},
		{"Hello", 7, "Hello", ""},
	} {
		a, b := cut(env.str, env.cp)
		ast := assert.New(t)
		ast.Equalf(env.a, a, "%+v", env)
		ast.Equalf(env.b, b, "%+v", env)
	}
}

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
		a.Equalf(env.result, result, "%+v", env)
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
		a.Equalf(env.result, env.counter, "%+v", env)
	}
}
