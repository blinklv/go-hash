// hash_test.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-12-04
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-12-09

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestQueue(t *testing.T) {
	for _, env := range []struct {
		input []*node
	}{
		{
			input: []*node{
				&node{i: 0},
				&node{i: 1},
				&node{i: 2},
				&node{i: 3},
				&node{i: 4},
				&node{i: 5},
				&node{i: 6},
				&node{i: 7},
				&node{i: 8},
				&node{i: 9},
			},
		},
		{
			input: []*node{
				&node{i: 9},
				&node{i: 8},
				&node{i: 7},
				&node{i: 6},
				&node{i: 5},
				&node{i: 4},
				&node{i: 3},
				&node{i: 2},
				&node{i: 1},
				&node{i: 0},
			},
		},
		{
			input: []*node{
				&node{i: 0},
				&node{i: 4},
				&node{i: 2},
				&node{i: 5},
				&node{i: 1},
				&node{i: 3},
				&node{i: 8},
				&node{i: 6},
				&node{i: 9},
				&node{i: 7},
			},
		},
	} {
		a := assert.New(t)
		i := 0
		for n := range queue(toInput(env.input)) {
			a.Equalf(n.i, i, "%+v", env)
			i++
		}
	}
}

func TestDisplay(t *testing.T) {
	for _, env := range []struct {
		input       []*node
		sumSize     int
		stdout      *bytes.Buffer
		result      []string
		errorExists bool
	}{
		{
			input:       []*node{},
			stdout:      &bytes.Buffer{},
			result:      []string{},
			errorExists: false,
		},
		{
			input: []*node{
				&node{path: "/hello/world", sum: []byte{0x12, 0x34, 0x56}},
				&node{path: "/foo/bar", sum: []byte{0xab, 0xcd, 0xef}},
				&node{path: "/hello/bar", sum: []byte{0x12, 0xcd, 0x56}},
			},
			stdout: &bytes.Buffer{},
			result: []string{
				"123456  /hello/world\n",
				"abcdef  /foo/bar\n",
				"12cd56  /hello/bar\n",
			},
			errorExists: false,
		},
		{
			input: []*node{
				&node{path: "/hello/world", sum: []byte{0x12, 0x34, 0x56}},
				&node{path: "/foo/bar", sum: []byte{0xab, 0xcd, 0xef}},
				&node{path: "/hello/bar", sum: []byte{0x12, 0xcd, 0x56}},
				&node{path: "/error", err: errorf("something wrong!")},
			},
			sumSize: 3,
			stdout:  &bytes.Buffer{},
			result: []string{
				"123456  /hello/world\n",
				"abcdef  /foo/bar\n",
				"12cd56  /hello/bar\n",
				"ERROR:  /error\n",
				" somet\n",
				"hing w\n",
				"rong! \n",
			},
			errorExists: true,
		},
	} {
		sumSize = env.sumSize
		stdout = env.stdout
		display(toInput(env.input))

		a := assert.New(t)
		a.Equalf(strings.Join(env.result, ""), env.stdout.String(), "%+v", env)
		a.Equalf(env.errorExists, errorExists, "%+v", env)
	}
}

func toInput(nodes []*node) chan *node {
	input := make(chan *node)
	go func() {
		for _, n := range nodes {
			input <- n
		}
		close(input)
	}()
	return input
}

func TestNodeNodes(t *testing.T) {
	for _, env := range []struct {
		n      node
		names  []string
		result []*node
	}{
		{
			node{path: "", depth: -1},
			[]string{},
			[]*node{},
		},
		{
			node{path: "", depth: -1},
			[]string{"foo", "bar", "/hello", "Apollo"},
			[]*node{
				&node{path: "foo", depth: 0},
				&node{path: "bar", depth: 0},
				&node{path: "Apollo", depth: 0},
				&node{path: "/hello", depth: 0},
			},
		},
		{
			node{path: "./foo/bar", depth: 10},
			[]string{"a/b/c", "Foo", "Hello World", ".", "/What/do/you"},
			[]*node{
				&node{path: "foo/bar/a/b/c", depth: 11},
				&node{path: "foo/bar/Hello World", depth: 11},
				&node{path: "foo/bar/Foo", depth: 11},
				&node{path: "foo/bar/What/do/you", depth: 11},
				&node{path: "foo/bar", depth: 11},
			},
		},
	} {
		nodes := env.n.nodes(env.names)
		a := assert.New(t)
		for i, n := range nodes {
			a.Equalf(env.result[i].path, n.path, "%+v", env)
			a.Equalf(env.n.depth+1, n.depth, "%+v", env)
		}
	}
}

func TestNodeFilename(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "hello.*.txt")
	defer os.Remove(tmpfile.Name())
	fileinfo, _ := tmpfile.Stat()

	for _, env := range []struct {
		n      node
		result string
	}{
		{
			n:      node{FileInfo: fileinfo, path: "/foo/bar"},
			result: fileinfo.Name(),
		},
		{
			n:      node{path: "/foo/bar"},
			result: "bar",
		},
		{
			n:      node{},
			result: "-",
		},
	} {
		a := assert.New(t)
		a.Equalf(env.result, env.n.filename(), "%+v", env)
	}
}

func TestNodeString(t *testing.T) {
	for _, env := range []struct {
		n        node
		filename bool
		sumSize  int
		result   string
	}{
		{
			n:        node{path: "/foo/bar", sum: []byte{0xab, 0x12, 0x33}, err: nil},
			filename: true,
			sumSize:  4,
			result:   "ab1233  /foo/bar",
		},
		{
			n:        node{path: "/foo/bar", sum: []byte{0xab, 0x12, 0x33}, err: nil},
			filename: false,
			sumSize:  4,
			result:   "ab1233",
		},
		{
			n: node{
				path: "/foo/bar",
				sum:  []byte{},
				err:  errorf("What do you want to do? Are you kidding me?"),
			},
			filename: false,
			sumSize:  4,
			result: strings.Join(
				[]string{
					"ERROR: W  /foo/bar",
					"hat do y",
					"ou want ",
					"to do? A",
					"re you k",
					"idding m",
					"e?      ",
				},
				"\n",
			),
		},
		{
			n: node{
				path: "/foo/bar",
				sum:  []byte{},
				err:  errorf("What do you want to do? Are you kidding me?"),
			},
			filename: false,
			sumSize:  8,
			result: strings.Join(
				[]string{
					"ERROR: What do y  /foo/bar",
					"ou want to do? A",
					"re you kidding m",
					"e?              ",
				},
				"\n",
			),
		},
	} {
		*_filename = env.filename
		sumSize = env.sumSize
		a := assert.New(t)
		a.Equalf(env.result, env.n.String(), "%+v", env)
	}
}

func TestSecretKeyInit(t *testing.T) {
	for _, env := range []struct {
		str string
		ok  bool
		key *secretKey
	}{
		{"what's wrong with you?", false, nil},
		{"binary:tmp.txt", true, &secretKey{"binary", "tmp.txt"}},
		{"hex:abcd1234", true, &secretKey{"hex", "abcd1234"}},
	} {
		key, err := (&secretKey{}).init(env.str)
		a := assert.New(t)
		a.Equalf(env.ok, err == nil, "%+v", env)
		a.Equalf(env.key, key, "%+v", env)
	}
}

func TestSecretKeyDecode(t *testing.T) {
	content := []byte("Hello, World!")
	tmpfile, _ := ioutil.TempFile("", "hello.*.txt")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)
	tmpfile.Close()

	b64Data := base64.StdEncoding.EncodeToString(content)
	hexData := hex.EncodeToString(content)

	for _, env := range []struct {
		key secretKey
		ok  bool
		b   []byte
	}{
		{secretKey{}, false, nil},
		{secretKey{scheme: "foo"}, false, nil},
		{secretKey{scheme: "binary", data: ""}, false, nil},
		{secretKey{scheme: "binary", data: tmpfile.Name()}, true, content},
		{secretKey{scheme: "base64", data: "Are you OK?"}, false, []byte{}},
		{secretKey{scheme: "base64", data: b64Data}, true, content},
		{secretKey{scheme: "hex", data: "Foo"}, false, []byte{}},
		{secretKey{scheme: "hex", data: hexData}, true, content},
	} {
		data, err := env.key.decode()
		a := assert.New(t)
		a.Equalf(env.ok, err == nil, "%+v", env)
		a.Equalf(env.b, data, "%+v", env)
	}
}

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
		{
			[]string{"a", "b", "foo", "H", "car", ".", "/", "hello", "world"},
			[]string{"world", "hello", "foo", "car", "b", "a", "H", "/", "."},
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
