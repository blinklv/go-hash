// hash.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-10-23
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-10-23

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
	"runtime"
)

// App binary version.
const binary = "1.0.0"

var (
	_algo     = flag.String("algo", "md5", "")
	_filename = flag.Bool("filename", true, "")
	_depth    = flag.Int("depth", 1, "")
	_version  = flag.Bool("version", false, "")
	_help     = flag.Bool("help", false, "")
)

func main() {
	parse_arg()
	fprintf(stdout, "%v", files)
	return
}

// When we call the digester function, we create a new hash.Hash instance to compute
// the digest of a file. Why don't we use the only one global hash.Hash instance?
// Because some hash.Hash implementations are not concurrent safe. If there're multiple
// digester goroutines use a global hash.Hash instance simultaneously, some exceptions
// maybe happen.
var creator factory

var files []string

// Parse the command arguments.
func parse_arg() {
	flag.Usage = func() { usage(stderr) } // overwrite the default usage.
	flag.Parse()

	if *_version {
		version(stdout)
		exit(nil)
	}

	if *_help {
		usage(stdout)
		exit(nil)
	}

	if creator = factories[*_algo]; creator == nil {
		exit(errorf("unknown hash algorithm '%s'", *_algo))
	}

	if files = flag.Args(); len(files) == 0 {
		exit(errorf("there is no file"))
	}

	return
}

// Output version information.
func version(w io.Writer) {
	fprintf(w, "%s v%s (built w/%s)\n", "go-hash", binary, runtime.Version())
}

// Output usage information.
func usage(w io.Writer) {
	msgs := []string{
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
		"       -version  - control whether to display version information. (default: false)\n",
		"\n",
		"       -help     - control whether to display usage information. (defualt: false)\n",
		"\n",
		"       file      - the objective file of the hash algorithm. If its type is directory,\n",
		"                   computing digests of all files in this directory recursively.\n",
		"\n",
	}
	for _, msg := range msgs {
		fprintf(w, msg)
	}
}

// Exit the process. If the 'e' parameter is not nil, print the error
// message and display usage information.
func exit(e error) {
	if e != nil {
		fprintf(stderr, "ERROR: %s\n", e)
		usage(stderr)
		os.Exit(1)
	}
	os.Exit(0)
}

// Rename some functions and objects to simplify my codes.
var (
	sprintf = fmt.Sprintf
	errorf  = fmt.Errorf
	fprintf = fmt.Fprintf
	stdout  = os.Stdout
	stderr  = os.Stderr
)

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
