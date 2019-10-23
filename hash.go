// hash.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2018-01-17
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
	"errors"
	"flag"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"syscall"
)

var (
	_filename = flag.Bool("filename", true, "")
	_depth    = flag.Int("depth", 1, "")
	_version  = flag.Bool("version", false, "")
	_help     = flag.Bool("help", false, "")
)

func parseArg() {
	flag.Parse()
	if *_version {
		version()
	} else if *_help {
		exit(0, "")
	} else {
	}
}

// Print version information and exit the process.
func version() {
	fmt.Printf("%s v%s (built w/%s)\n", "go-hash", binary, runtime.Version())
	os.Exit(0)
}

// Output usage information.
func usage(w io.Writer) {
	msgs := []string{
		"usage: go-hash [option] algorithm file...\n",
		"\n",
		"       algorithm - the hash algorithm for computing the digest of files.\n",
		"                   Its values can be one in the following list (default: md5)\n",
		"\n",
		"                   md5, sha1, sha224, sha256, sha384, sha512, sha512/224\n",
		"                   sha512/256, fnv32, fnv32a, fnv64, fnv64a, fnv128, fnv128a\n",
		"\n",
		"       file      - the objective file of the hash algorithm. If its type is directory,\n",
		"                   computing digests of all files in this directory recursively.\n",
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
	}
	for _, msg := range msgs {
		fmt.Fprintf(w, msg)
	}
}

func main() {
	parseArg()

	var (
		err        error
		m          map[string][]byte
		exit, done = make(chan struct{}), make(chan struct{})
		sigch      = make(chan os.Signal, 8)
	)

	go func() {
		m, err = hashAll(os.Args[2:], exit)
		close(done)
	}()

	// There're two cases will cause the process exit. The first case
	// is trival, computing digests of all files has done. The second
	// case is triggered by some OS signals, it will make the process
	// exits ahead of time.
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigch:
		close(exit)
		// It will check the value of 'err' variable, so we need to
		// wait for 'hasAll' function has done.
		<-done
	case <-done:
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s!\n", err)
		os.Exit(1)
	}

	paths := make([]string, 0, len(m))
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}

	return
}

// When we call the digester function, we create a new hash.Hash instance to compute
// the digest of a file. Why don't we use the only one global hash.Hash instance?
// Because some hash.Hash implementations are not concurrent safe. If there're multiple
// digester goroutines use a global hash.Hash instance simultaneously, some exceptions
// maybe happen.
var creator factory

// Parse the command arguments.
func parseArg() {
	if len(os.Args) == 1 {
		exit(1, "arguments are empty")
	}

	if creator = factories[os.Args[1]]; creator == nil {
		switch os.Args[1] {
		case "version":
			version()
		case "help":
			exit(0, "")
		default:
			exit(1, fmt.Sprintf("invalid option (%s)", os.Args[1]))
		}
	} else if len(os.Args) == 2 {
		// There must exist one file at least when the first argument
		// is the name of a hash algorithm.
		exit(1, "file arguments are empty")
	}

	return
}

// Print error, usage information and exit the process.
func exit(code int, msg string) {
	w := os.Stdout
	if code != 0 {
		w = os.Stderr
		fmt.Fprintf(w, "error: %s!\n\n", msg)
	}
	usage(w)
	os.Exit(code)
}

const binary = "1.0.0"

// Print version information and exit the process.
func version() {
	fmt.Printf("%s v%s (built w/%s)\n", "go-hash", binary, runtime.Version())
	os.Exit(0)
}

type result struct {
	path string
	sum  []byte
	err  error
}

func hashAll(roots []string, exit chan struct{}) (map[string][]byte, error) {
	results, errc := sum(roots, exit)
	m := make(map[string][]byte)
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}

	for err := range errc {
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// If we allocate a goroutine for each path immediately, it will cost resources heavy
// when there're so many big files. So we should limit the number of goroutines computing
// digest to reduce side effects for OS.
const numDigester = 16

func sum(roots []string, exit <-chan struct{}) (<-chan result, <-chan error) {
	results := make(chan result)
	paths, errc := walk(roots, exit)

	wg := &sync.WaitGroup{}
	wg.Add(numDigester)
	for i := 0; i < numDigester; i++ {
		go func() {
			digester(paths, results, exit)
			wg.Done()
		}()
	}

	go func() {
		// We must ensure that no one writes to 'results' channel before closing it.
		wg.Wait()
		close(results)
	}()

	return results, errc
}

// Emits the paths for regular files in the tree.
func walk(roots []string, exit <-chan struct{}) (<-chan string, <-chan error) {
	paths, errc := make(chan string), make(chan error, len(roots))
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(roots))

		for _, root := range roots {
			go func(root string) {
				// No select needed for this send, since errc is buffered.
				errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.Mode().IsRegular() {
						return nil
					}
					select {
					case paths <- path:
					case <-exit:
						return errors.New("walk canceled")
					}
					return nil
				})
				wg.Done()
			}(root)
		}

		go func() {
			wg.Wait()

			// Close the paths and errc channel after Walk returns. If we don't do this operation,
			// 'digester' won't exit, then 'hashAll' won't exit too.
			close(paths)
			close(errc)
		}()
	}()
	return paths, errc
}

func digester(paths <-chan string, results chan<- result, exit <-chan struct{}) {
	// Some hash.Hash implementations are not concurrent safe, so we need to
	// create a new instance for a single digester goroutine.
	h := creator()

	for path := range paths {
		data, err := ioutil.ReadFile(path)
		h.Reset() // Don't forget this operation.
		h.Write(data)

		select {
		case results <- result{path, h.Sum(nil), err}:
		case <-exit:
			return
		}
	}
}

// Output version information.
func version(w io.Writer) {
	fprintf(stdout, "%s v%s (built w/%s)\n", "go-hash", binary, runtime.Version())
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
