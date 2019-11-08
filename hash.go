// hash.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2019-10-23
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2019-11-08

// A simple command tool to calculate the digest value of files. It supports some
// primary Message-Digest Hash algorithms, like MD5, FNV family, and SHA family.
package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
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
	"strings"
	"sync"
	"syscall"
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

// Number of hash sum bytes.
var sumSize int

// Keyed-Hash Message Authentication Code (HMAC) sign key.
var hmacKey []byte

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

// keydecs variable specifies all key formats.
var keydecs = map[string]keydec{
	"binary": keydecBinary,
	"base64": keydecBase64,
	"hex":    keydecHex,
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
	"       -hmac_key - HMAC secret key. It will compute hash-based message authentication codes\n",
	"                   instead of digests when you specify a legal key. A key should meet the\n",
	"                   requirements: 'encoding-scheme':'data'. The combinations you can select:\n",
	"\n",
	"                       'binary':'path of the secret key file'\n",
	"                       'base64':'standard base64 encoded string'\n",
	"                          'hex':'hex encoded string'\n",
	"\n",
	"       -version  - control whether to display version information. (default: false)\n",
	"\n",
	"       -help     - control whether to display usage information. (defualt: false)\n",
	"\n",
	"       file      - the objective file of the hash algorithm. If its type is directory,\n",
	"                   computing digests of all files in this directory recursively. This\n",
	"                   tool will read from the stdin when no file specified.\n",
	"\n",
}

// Command-Line options.
var (
	_algo     = flag.String("algo", "md5", "")
	_filename = flag.Bool("filename", true, "")
	_depth    = flag.Int("depth", 1, "")
	_all      = flag.Bool("all", false, "")
	_hmac_key = flag.String("hmac_key", "", "")
	_version  = flag.Bool("version", false, "")
	_help     = flag.Bool("help", false, "")
)

/* Main Functions */

func main() {
	var (
		exit, done = make(trigger), make(trigger)
		signals    = make(chan os.Signal, 8)
	)

	go func() {
		display(queue(digester(walk(exit, parse_arg()))))
		close(done)
	}()

	// There're two cases will cause the process exit. The first case
	// is trival, computing digests of all files has done. The second
	// case is triggered by some OS signals, it will make the process
	// exits ahead of time.
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signals:
		close(exit)
		<-done
	case <-done:
	}
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
	sumSize = (creator()).Size()

	if *_hmac_key != "" {
		var k = &key{}
		err := k.initialize(*_hmac_key)
		if err != nil {
			exit(err)
		}

		var dec keydec
		if dec = keydecs[k.scheme]; dec == nil {
			exit(errorf("unknown key scheme (%s)", k.scheme))
		}

		if hmacKey, err = dec(k.data); err != nil {
			exit(errorf("hmac key (%s) is illegal: %s", *_hmac_key, err))
		}
		creator = factoryHMAC(creator).normalize()
	}

	return flag.Args() // Root files to be processed.
}

// Traverse a directory tree in pre-order and push nodes to the output channel.
func walk(exit trigger, roots []string) (output chan *node) {
	output = make(chan *node)
	go func() {
		// Initialize the node stack. Cause the depth of root nodes
		// is zero; the depth of their parent is -1.
		var S []*node
		if len(roots) > 0 {
			S = (&node{depth: -1}).nodes(roots)
		} else {
			S = []*node{&node{}}
		}

		// Iterate the node stack.
		var top *node
		for i := 0; len(S) > 0; {
			top, S = S[len(S)-1], S[:len(S)-1]        // Pop the top node.
			if !(*_all) && isHidden(top.filename()) { // Skip hidden files.
				continue
			}

			if top.err == nil && top.depth < *_depth {
				if children := top.children(); len(children) > 0 {
					S = append(S, children...)
				}
			}

			select {
			case output <- top.mark(i):
				i++
			case <-exit:
				goto end
			}
		}
	end:
		close(output)
	}()
	return output
}

// Get the file information from the input channel and compute its digest,
// then push the result to the output channel.
func digester(input chan *node) (output chan *node) {
	output = make(chan *node, numDigester)
	wg := &sync.WaitGroup{}

	for i := 0; i < numDigester; i++ {
		wg.Add(1)
		go func() {
			// Some hash.Hash implementations are not concurrent safe, so we need to
			// create a new instance for a single digester goroutine.
			h := creator()
			for n := range input {
				h.Reset() // NOTE: Don't forget this operation.

				data, err := n.read()
				h.Write(data)
				n.sum, n.err = h.Sum(nil), err
				output <- n
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// Queue output results by walking sequence. (Running in a separate goroutine)
func queue(input chan *node) (output chan *node) {
	output = make(chan *node)
	go func() {
		var (
			next  int                   // Mark the next node which we want to output.
			cache = make(map[int]*node) // Caches nodes that haven't been outputed.
		)

		for n := range input {
			if n.i == next {
				output <- n
				next++

				for n = cache[next]; n != nil; n = cache[next] {
					delete(cache, next)
					output <- n
					next++
				}

			} else {
				cache[n.i] = n
			}
		}
		close(output)
	}()
	return output
}

// Print digest of files to the standard output.
func display(input chan *node) {
	for n := range input {
		if !n.isdir() {
			fprintf(stdout, "%s\n", n)
		}
	}
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
	if n.isdir() {
		var names []string
		if names, n.err = readdir(n.path); n.err != nil {
			return nil
		}
		return n.nodes(names)
	}
	return nil
}

// Convert multiple filenames to the corresponded nodes based on the current node.
func (n *node) nodes(names []string) []*node {
	var ns = make([]*node, 0, len(names))
	for _, name := range rsort(names) {
		ns = append(ns, (&node{
			depth: n.depth + 1,
		}).initialize(filepath.Join(n.path, name)))
	}
	return ns
}

// isdir() checks whether the node describes a directory.
func (n *node) isdir() bool {
	if n.FileInfo != nil {
		return n.IsDir()
	}
	return false
}

// filename() returns the file name of the node.
func (n *node) filename() string {
	if n.FileInfo != nil {
		return n.Name()
	}

	// NOTE: filepath.Base returns "." when the path is empty, which will
	// cause the standard input to be skipped in normal case (_all flag is
	// unset). So I use the _path() method instead of the path field.
	return filepath.Base(n._path() /* not path */)
}

// read() reads from the file or the standard input until an error or EOF
// and returns the data it read.
func (n *node) read() ([]byte, error) {
	if n.path != "" {
		return ioutil.ReadFile(n.path)
	}
	return ioutil.ReadAll(os.Stdin)
}

// _path() returns "-" instead of an empty string when the path is empty.
func (n *node) _path() string {
	if n.path != "" {
		return n.path
	}
	return "-" // Represents the standard input (stdin).
}

// String returns the string form of a node.
func (n *node) String() string {
	if n.err == nil && *_filename {
		return sprintf("%x  %s", n.sum, n._path())
	} else if n.err == nil && !(*_filename) {
		return sprintf("%x", n.sum)
	} else { // The file name can't be skipped when something wrong.
		var (
			rest  = sprintf("ERROR: %s", n.err)
			line  string
			lines = make([]string, 0, 1)
		)

		for rest != "" {
			line, rest = cut(rest, 2*sumSize)
			// The first line contains the file name.
			if len(lines) == 0 {
				line = sprintf(sprintf("%%-%ds  %%s", 2*sumSize), line, n._path())
			}
			lines = append(lines, line)
		}
		return strings.Join(lines, "\n")
	}
}

// factory specifices how to create a hash.Hash instance.
type factory func() hash.Hash

// factory32 specifies how to create a hash.Hash32 instance.
type factory32 func() hash.Hash32

// Converts a factory32 instance to the corresponded factory instance.
func (f32 factory32) normalize() factory {
	return func() hash.Hash { return f32() }
}

// factory64 specifies how to create a hash.Hash64 instance.
type factory64 func() hash.Hash64

// Converts a factory64 instance to the corresponded factory instance.
func (f64 factory64) normalize() factory {
	return func() hash.Hash { return f64() }
}

// factoryHMAC specifies how to create a HMAC hash.Hash instance.
type factoryHMAC func() hash.Hash

// Converts a factoryHMAC instance to the corresponded factory instance.
func (fh factoryHMAC) normalize() factory {
	return func() hash.Hash { return hmac.New(fh, hmacKey) }
}

// Inner representations of all *_key options (eg. hmac_key)
type key struct {
	scheme string
	data   string
}

// Initialize a key from a *_key option (eg. hmac_key)
func (k *key) initialize(str string) error {
	strs := strings.SplitN(str, ":", 2)
	if len(strs) != 2 {
		return errorf("invalid key '%s'", str)
	}
	k.scheme, k.data = strs[0], strs[1]
	return nil
}

// keydec (Secret Key Decoder) specifies how to decode a key of a particular format.
type keydec func(string) ([]byte, error)

// Decode a key from a normal binary file.
var keydecBinary = ioutil.ReadFile

// Decode a key from a base64 encoded string.
var keydecBase64 = base64.StdEncoding.DecodeString

// Decode a key from a hex encoded string.
var keydecHex = hex.DecodeString

/* Auxiliary Functions */

// The only reason I rename the following functions is simplify my codes.
var (
	sprintf = fmt.Sprintf
	errorf  = fmt.Errorf
	fprintf = fmt.Fprintf
	stdout  = os.Stdout
	stderr  = os.Stderr
)

// rsort (Reverse Sort) sorts a slice of strings in decreasing alphabetical order.
func rsort(strs []string) []string {
	sort.Sort(sort.Reverse(sort.StringSlice(strs)))
	return strs
}

// Reads the directory named by dirname and returns a list of entries name.
func readdir(dirname string) ([]string, error) {
	dir, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return dir.Readdirnames(-1)
}

// Cut a string into two parts. If the size of a string is not greater than the
// cut point, the second part will be empty.
func cut(str string, cp int /* cut point */) (a, b string) {
	if len(str) > cp {
		return str[:cp], str[cp:]
	}
	return str, ""
}
