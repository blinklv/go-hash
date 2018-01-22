// hash.go
//
// Author: blinklv <blinklv@icloud.com>
// Create Time: 2018-01-17
// Maintainer: blinklv <blinklv@icloud.com>
// Last Change: 2018-01-22

// A simple command tool to calculate the HASH value of files. It supports
// some mainstream HASH algorithms, like MD5, FNV family and SHA family.
package main

import (
	"flag"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sync"
)

var h hash.Hash

func main() {
}

type result struct {
	path string
	sum  []byte
	err  error
}

func hashAll(root string, done chan struct{}) (map[string][]byte, error) {
	results, errc := sum(root, done)
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

func sum(roots []string, done <-chan struct{}) (<-chan result, <-chan error) {
	results := make(chan result)
	paths, errc := walk(roots, done)

	wg := &sync.WaitGroup{}
	wg.Add(numDigester)
	for i := 0; i < numDigester; i++ {
		go func() {
			digester(paths, results, done)
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
func walk(roots []string, done <-chan struct{}) (<-chan string, <-chan error) {
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
					case <-done:
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

func digester(paths <-chan string, results chan<- result, done <-chan struct{}) {
	for path := range paths {
		data, err := ioutil.ReadFile(path)
		select {
		case results <- result{path, h.Sum(data), err}:
		case <-done:
			return
		}
	}
}
