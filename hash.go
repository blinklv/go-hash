// hash.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-10-23
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-10-31

// A simple command tool to calculate the digest value of files. It supports some
// primary Message-Digest Hash algorithms, like MD5, FNV family, and SHA family.
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

/* Global Constants and Variables */

const binary = "1.0.0" // App binary version.

// If we allocate a goroutine for each path immediately, it will cost resources heavy
// when there're so many big files. So we should limit the number of goroutines computing
// digest to reduce side effects for OS.
const numDigester = 16

// When we call the digester function, we create a new hash.Hash instance to compute
// the digest of a file. Why don't we use the only one global hash.Hash instance?
// Because some hash.Hash implementations are not concurrent safe. If there're multiple
// digester goroutines use a global hash.Hash instance simultaneously, some exceptions
// maybe happen.
var creator factory

// factories variable specifies all HASH algorithms supported by this tool.
var factories = map[string]factory{
	"md5":        md5.New,
	"sha1":       sha1.New,
	"sha224":     sha256.New224,
	"sha256":     sha256.New,
	"sha384":     sha512.New384,
	"sha512":     sha512.New,
	"sha512/224": sha512.New512_224,
	"sha512/256": sha512.New512_256,
	"fnv32":      (factory32(fnv.New32)).normalize(),
	"fnv32a":     (factory32(fnv.New32a)).normalize(),
	"fnv64":      (factory64(fnv.New64)).normalize(),
	"fnv64a":     (factory64(fnv.New64a)).normalize(),
	"fnv128":     fnv.New128,
	"fnv128a":    fnv.New128a,
}

// Help document.
var usages = []string{
	"usage: go-hash [option] file...\n",
	"\n",
	"       -algo     - the hash algorithm for computing the digest of files. (default: md5)\n",
	"                   Its values can be one in the following list:\n",
	"\n",
	"                   md5, sha1, sha224, sha256, sha384, sha512, sha512/224\n",
	"                   sha512/256, fnv32, fnv32a, fnv64, fnv64a, fnv128, fnv128a\n",
	"\n",
	"       -filename - control whether to display the corresponded filenames when outputing\n",
	"                   the digest of files. (default: true)\n",
	"\n",
	"       -depth    - control the recursive depth of searching directories. (default: 1)\n",
	"\n",
	"       -all      - control whether process hidden files. (default: false)\n",
	"\n",
	"       -version  - control whether to display version information. (default: false)\n",
	"\n",
	"       -help     - control whether to display usage information. (defualt: false)\n",
	"\n",
	"       file      - the objective file of the hash algorithm. If its type is directory,\n",
	"                   computing digests of all files in this directory recursively.\n",
	"\n",
}

// Command-Line options.
var (
	_algo     = flag.String("algo", "md5", "")
	_filename = flag.Bool("filename", true, "")
	_depth    = flag.Int("depth", 1, "")
	_all      = flag.Bool("all", false, "")
	_version  = flag.Bool("version", false, "")
	_help     = flag.Bool("help", false, "")
)

/* Main Functions */

func main() {
	parse_arg()
	return
}

// Parse the command arguments and return root files.
func parse_arg() []string {
	flag.Usage = func() { help(stderr) } // overwrite the default usage.
	flag.Parse()

	if *_version {
		version(stdout)
		exit(nil)
	}

	if *_help {
		help(stdout)
		exit(nil)
	}

	if creator = factories[*_algo]; creator == nil {
		exit(errorf("unknown hash algorithm '%s'", *_algo))
	}

	var roots []string // Root files to be processed.
	if roots = flag.Args(); len(roots) == 0 {
		exit(errorf("no input file specified"))
	}
	return roots
}

// Traverse a directory tree in preorder.
func walk(exit trigger, roots []string) chan *node {
	nodes := make(chan *node, numDigester)
	go func() {
		// Initialize the node stack.
		rsort(roots)
		S := make([]*node, 0, len(roots))
		for _, root := range roots {
			S = append(S, (&node{}).initialize(root))
		}

		// Iterate the node stack.
		var top *node
		for i := 0; len(S) > 0 && S[len(S)-1].depth <= *_depth; i++ {
			top, S = S[len(S)-1], S[:len(S)-1] // Pop the top node.
			if top.err == nil {
				if children := top.children(); len(children) > 0 {
					S = append(S, children...)
				}
			}
			nodes <- top.mark(i)
		}
	}()
	return nodes
}

// Output version information.
func version(w io.Writer) {
	fprintf(w, "%s v%s (built w/%s)\n", "go-hash", binary, runtime.Version())
}

// Exit the process. If the 'e' parameter is not nil, print the error
// message and display usage information.
func exit(e error) {
	if e != nil {
		fprintf(stderr, "ERROR: %s\n", e)
		help(stderr)
		os.Exit(1)
	}
	os.Exit(0)
}

// Output usage information.
func help(w io.Writer) {
	for _, usage := range usages {
		fprintf(w, usage)
	}
}

/* Auxiliary Structs and Methods */

type trigger chan struct{}

// Directory tree node.
type node struct {
	os.FileInfo
	path  string // Filepath
	i     int    // Walk sequence
	depth int    // Directory depth
	sum   []byte // Digest
	err   error
}

// Initialize a node by using its path and return itself. If there
// exists any problem, it will store the error to the 'err' field.
func (n *node) initialize(path string) *node {
	n.path = path
	n.FileInfo, n.err = os.Lstat(path)
	return n
}

// Using 'walk sequence' to mark a node has been traversed.
func (n *node) mark(i int) *node {
	n.i = i
	return n
}

// Return children node of a directory node. If there is something wrong,
// it will store the error to the 'err' field.
func (n *node) children() []*node {
	if n.IsDir() {
		var (
			f     *os.File
			names []string
		)

		f, n.err = os.Open(n.path)
		if n.err != nil {
			return nil
		}

		names, n.err = f.Readdirnames(-1)
		f.Close()
		if n.err != nil {
			return nil
		}

		rsort(names)
		children := make([]*node, 0, len(names))
		for _, name := range names {
			children = append(children, (&node{
				depth: n.depth + 1,
			}).initialize(join(n.path, name)))
		}
	}
	return nil
}

// factory specifices how to create a hash.Hash instance.
type factory func() hash.Hash

// factory32 specifies how to create a hash.Hash32 instance.
type factory32 func() hash.Hash32

// Converts a factory32 instance to the corresponded factory instance.
func (f32 factory32) normalize() factory {
	return func() hash.Hash { return f32() }
}

// factory64 specifies how to create a hash.Hasn64 instance.
type factory64 func() hash.Hash64

// Converts a factory64 instance to the corresponded factory instance.
func (f64 factory64) normalize() factory {
	return func() hash.Hash { return f64() }
}

/* Auxiliary Functions */

// The only reason I rename the following functions is simplify my codes.
var (
	sprintf = fmt.Sprintf
	errorf  = fmt.Errorf
	fprintf = fmt.Fprintf
	stdout  = os.Stdout
	stderr  = os.Stderr
	join    = filepath.Join
)

// rsort (Reverse Sort) sorts a slice of strings in decreasing alphabetical order.
func rsort(strs []string) {
	sort.Sort(sort.Reverse(sort.StringSlice(strs)))
}
