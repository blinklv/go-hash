# go-hash

[![Build Status](https://travis-ci.com/blinklv/go-hash.svg?branch=master)](https://travis-ci.com/blinklv/go-hash)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Version](https://img.shields.io/badge/version-1.0.0-green.svg)

**go-hash** is a simple command tool to calculate the digest of files. It supports some primary *Message-Digest Hash* algorithms, like [MD5][], [FNV][] family and [SHA][] family.

## Install

```bash
$ git clone https://github.com/blinklv/go-hash.git
$ cd go-hash && go install
```

## Usage

**Format**

```bash
go-hash [option] file...

   -algo     - the hash algorithm for computing the digest of files. (default: md5)
               Its values can be one in the following list:

               md5, sha1, sha224, sha256, sha384, sha512, sha512/224
               sha512/256, fnv32, fnv32a, fnv64, fnv64a, fnv128, fnv128a

   -filename - control whether to display the corresponded filenames when outputing
               the digest of files. (default: true)

   -depth    - control the recursive depth of searching directories. (default: 1)

   -all      - control whether process hidden files. (default: false)

   -hmac_key - HMAC secret key. It will compute hash-based message authentication codes
               instead of digests when you specify a legal key. A key should meet the
               requirements: 'encoding-scheme':'data'. The combinations you can select:

                   'binary':'path of the secret key file'
                   'base64':'standard base64 encoded string'
                      'hex':'hex encoded string'

   -version  - control whether to display version information. (default: false)

   -help     - control whether to display usage information. (defualt: false)

   file      - the objective file of the hash algorithm. If its type is directory,
               computing digests of all files in this directory recursively. This
               tool will read from the stdin when no file specified.
```

**Compute the digest of a single file**

- *Default*

```bash
$ go-hash LICENSE

f91b07d7eebf9380c2279ea572c6366a  LICENSE
```

- *Using a different hash algorithm*

```bash
$ go-hash -algo sha1 LICENSE

7d1c7ec803a19ea10069e0838d02aa778ba4f9bb  LICENSE
```

- *Do not display file name*

```bash
$ go-hash -filename=false LICENSE

f91b07d7eebf9380c2279ea572c6366a
```

- *From stdin*

```bash
$ cat LICENSE | go-hash

f91b07d7eebf9380c2279ea572c6366a  -
```

> **NOTE**: You can't read data from files and stdin at the same time.

- *With HMAC secret key*

```bash
$ go-hash -hmac_key='hex:0123456789abcdef' LICENSE

e0068df864b3ff7d748aa6861d216a76  LICENSE
```

**Compute the digests of multiple files**

```bash
$ go-hash hash.go hash_test.go is_hidden.go

5a7a0f610d372fff7cc949ce3b21c3ff  hash.go
5a276e39f1c659525455617e13507e80  hash_test.go
3d4ea5b1571409c98366983a7c01263a  is_hidden.go
```

**Compute the digests of a directory**

- *Hidden directory*

```bash
$ go-hash -all .git

3c53ae7a64d088b297fe3d8cbacc3406  .git/COMMIT_EDITMSG
deb816b71913842b6025cde7400862d2  .git/HEAD
3dd68f97cbaee6858b96a46b2de3eae5  .git/config
a0a7c3fff21f2aea3cfa1d0316dd816c  .git/description
9ac276273e12cf6c0d93d400cf305dd9  .git/index
49132b84108816a83a58a10f799ec9cc  .git/packed-refs
```

- *Include subdirectories*

```bash
$ go-hash -all -depth=2 .git

3c53ae7a64d088b297fe3d8cbacc3406  .git/COMMIT_EDITMSG
deb816b71913842b6025cde7400862d2  .git/HEAD
3dd68f97cbaee6858b96a46b2de3eae5  .git/config
a0a7c3fff21f2aea3cfa1d0316dd816c  .git/description
ce562e08d8098926a3862fc6e7905199  .git/hooks/applypatch-msg.sample
579a3c1e12a1e74a98169175fb913012  .git/hooks/commit-msg.sample
ecbb0cb5ffb7d773cd5b2407b210cc3b  .git/hooks/fsmonitor-watchman.sample
2b7ea5cee3c49ff53d41e00785eb974c  .git/hooks/post-update.sample
054f9ffb8bfe04a599751cc757226dda  .git/hooks/pre-applypatch.sample
e4db8c12ee125a8a085907b757359ef0  .git/hooks/pre-commit.sample
3c5989301dd4b949dfa1f43738a22819  .git/hooks/pre-push.sample
56e45f2bcbc8226d2b4200f7c46371bf  .git/hooks/pre-rebase.sample
2ad18ec82c20af7b5926ed9cea6aeedd  .git/hooks/pre-receive.sample
2b5c047bdb474555e1787db32b2d2fc5  .git/hooks/prepare-commit-msg.sample
517f14b9239689dff8bda3022ebd9004  .git/hooks/update.sample
9ac276273e12cf6c0d93d400cf305dd9  .git/index
036208b4a1ab4a235d75c181e685e5a3  .git/info/exclude
8cf658368300787b235dcb95d10069b6  .git/logs/HEAD
49132b84108816a83a58a10f799ec9cc  .git/packed-refs
```

**Compute the digests of the combination of files and directories**

```bash
$ go-hash -all -depth=2 *.go .git

3c53ae7a64d088b297fe3d8cbacc3406  .git/COMMIT_EDITMSG
deb816b71913842b6025cde7400862d2  .git/HEAD
3dd68f97cbaee6858b96a46b2de3eae5  .git/config
a0a7c3fff21f2aea3cfa1d0316dd816c  .git/description
ce562e08d8098926a3862fc6e7905199  .git/hooks/applypatch-msg.sample
579a3c1e12a1e74a98169175fb913012  .git/hooks/commit-msg.sample
ecbb0cb5ffb7d773cd5b2407b210cc3b  .git/hooks/fsmonitor-watchman.sample
2b7ea5cee3c49ff53d41e00785eb974c  .git/hooks/post-update.sample
054f9ffb8bfe04a599751cc757226dda  .git/hooks/pre-applypatch.sample
e4db8c12ee125a8a085907b757359ef0  .git/hooks/pre-commit.sample
3c5989301dd4b949dfa1f43738a22819  .git/hooks/pre-push.sample
56e45f2bcbc8226d2b4200f7c46371bf  .git/hooks/pre-rebase.sample
2ad18ec82c20af7b5926ed9cea6aeedd  .git/hooks/pre-receive.sample
2b5c047bdb474555e1787db32b2d2fc5  .git/hooks/prepare-commit-msg.sample
517f14b9239689dff8bda3022ebd9004  .git/hooks/update.sample
9ac276273e12cf6c0d93d400cf305dd9  .git/index
036208b4a1ab4a235d75c181e685e5a3  .git/info/exclude
8cf658368300787b235dcb95d10069b6  .git/logs/HEAD
49132b84108816a83a58a10f799ec9cc  .git/packed-refs
5a7a0f610d372fff7cc949ce3b21c3ff  hash.go
5a276e39f1c659525455617e13507e80  hash_test.go
3d4ea5b1571409c98366983a7c01263a  is_hidden.go
ecfd959ec7ee3cee14bc7089147ac52e  is_hidden_windows.go
```

[MD5]: https://en.wikipedia.org/wiki/MD5
[FNV]: https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
[SHA]: https://en.wikipedia.org/wiki/Secure_Hash_Algorithms
