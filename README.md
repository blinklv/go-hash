# go-hash

![Building](https://img.shields.io/badge/building-passing-green.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Version](https://img.shields.io/badge/version-1.0.0-green.svg)

**go-hash** is simple command tool to calculate the digest of files. It supports some mainstream algorithms, like [MD5][], [FNV][] family and [SHA][] family. 

## Install

```bash
$ go get github.com/blinklv/go-hash
$ go install
```

## Usage

**Compute the digest of a single file**

```bash
$ go-hash md5 LICENSE

f91b07d7eebf9380c2279ea572c6366a  LICENSE
```

**Compute the digests of multiple files**

```bash
$ go-hash fnv128a LICENSE README.md hash.go

0aef0e483a88ca81138ef02a81186636  LICENSE
b2bcd83298747f4ad9668d96a059e3c9  README.md
ea355f219bab32cc7a3aedb23070c152  hash.go
```

**Compute the digests of all files in a directory**

```bash
$ go-hash sha256 .

426133479b701a56f5abef378f6d51fc7f1dc3d74d0387af886f71d43c19f0e5  .jscsrc
8a0f5e9593e902f59bec44cbba58114f529aa2783f9d517d2da7f3886cb49d70  .jshintrc
0e3fcef1bd69eb73f0cc8ef56485e613d3664fc43e960f7a4d5355c7c0c3a47a  affix.js
af26e2d2d3b44216f53d6b1e0597b04f82ab4299b0e89638b750ede1bcf6fc17  alert.js
85ab1ee20edff94e8e96425b77510c14017fbae956e4c11913651db0f1218a13  button.js
7fcb97936241bb603ec42136e7fd7f277e06bd12adebabcf878026bbca1fadf0  carousel.js
91cfa7a40d2a9b731365268eef2bebf108888d3386bac5260eae03443eda5f18  collapse.js
93ba2b87b9e61844b7f808cdac165ac2bf031bbad9a5e1d2f9d83b6db6b842a4  dropdown.js
5dcdb087eb65e0122c7474c4d729f3c28568ec7f4fde3b7f63596de689ad9a20  modal.js
8a14cd5f4e1f6105ff3a6e9c1a7c5dec3a13095d610f47d5c8bf7833eb403df7  popover.js
bfa5d8fc5b6a51f408fb1ba27342dc50a54b69e2adddc01eaac5e0bbacf720a6  scrollspy.js
4d7d9559fe2f8df53c5e015dba67ec75fcf9f94b10cdad625f1e0f223f4db47b  tab.js
9758b226ca51d2319912cbb2e28036ea8a61f506e0ac8653da849559b34e36ad  tests/README.md
70df0e3997593867089f6fb342c1bbcee6ee95661d442eb9b89f45e346e0406d  tests/index.html
99572a133fe0aa6e5c0a15b981952eef12f4be42e34de0c0f7f43994a1df55f4  tests/unit/.jshintrc
3e2d597f1470ab9df0d4a7881e3388c4c7f9c069a1b015d54f7295397846c1ff  tests/unit/affix.js
893943eaf59f093b993667c5c2a993d6d2aa285ea2867412ced1d03c2db89329  tests/unit/alert.js
3c3ba764bbcfe7fde855d905ab79ac39193f4ea779a8e9c06ce1121af8950a32  tests/unit/button.js
145cb89bef2215130a20de5ed8fd3a6848b96e87aa901d071d6c1fed5264353a  tests/unit/carousel.js
4f31c4d5d101624c6cf605f0abd68f37e566396b10571c0dfe489e29319bbe11  tests/unit/collapse.js
237792141238df5e1599708489d2d08ce4b7db9a52d097d3e8103bc9cd879001  tests/unit/dropdown.js
d66a938ab781599c0d36fdf0e091afa521d0c9029498d531ccc41dd3378c7043  tests/unit/modal.js
f6ab3dacd25564dfd71756e935076570bf52cd67e502b7f4e8019ded92047016  tests/unit/phantom.js
fee082e2a2b02d55c0fe6971a0b8a92cc72d4d072872d4ad24e2536ef04672d6  tests/unit/popover.js
53a5e3aab6f43bd14f5d3e5f991ab4dab75b573c0eaedd074fe22974994ce472  tests/unit/scrollspy.js
eb8d656566bf36997c9a69d2b9b18f581dc7867138ad98c24dd9faa089e57a49  tests/unit/tab.js
52321b7bcf76efa3d282d22d38a72f8b51fcd83cde369f967d9579a83ea0267e  tests/unit/tooltip.js
ecb916133a9376911f10bc5c659952eb0031e457f5df367cde560edbfba38fb8  tests/vendor/jquery.min.js
7ca25a5c17b77a63926bafbbb2893eba4bd9e8f01383f640d7a89784c43c19c2  tests/vendor/qunit.css
7bfc59ee5478e6b08422d5224f1c2bc1c66ca8696ee1ca196ef75a0f29ebbfba  tests/vendor/qunit.js
3fe72c5e6a7b6e0a0f51e69fb45bb7750dcbb52265c60a62128e2dba7d3a5ce4  tests/visual/affix-with-sticky-footer.html
729924f8e0cc71d93a4a9a0727fc632356ce182fb284919efca18ab68fda8b2b  tests/visual/affix.html
0862b81b2034375a7f633a67b1d3c9ea0381060d75a42254452e34c636e0e141  tests/visual/alert.html
2a33cc23dcd21255e23124570eb68dcdb4499d558d7bb30ac5a50dbdeeade7fa  tests/visual/button.html
7054ce371f0d52f37241da86f08c70aaa43d0f4db61c008a624ef848444a3394  tests/visual/carousel.html
c51d07195a52488da7c5aa95257dd67b32ded4577ffd2a1bcdb66e03ccd1f749  tests/visual/collapse.html
47e769ebcc8a94d980cade0e62df5253cfdd1de1f870ec74ee6a957a9a1fd8ca  tests/visual/dropdown.html
11ae29ca0e5916cb26209e303acce6166419a636b57934ee2849b328082eeb15  tests/visual/modal.html
94c089635ac78bde7e4ba6ce902eb84ca966e1bfbdc856fa27896069e4a5fd59  tests/visual/popover.html
c349d0741945ebafa8391f29d0981acab7858d08191d8daedb9d20187b946064  tests/visual/scrollspy.html
69f85927acb03c927fa32da51bba444db83936e26b7cf43d7009922ba2a04960  tests/visual/tab.html
795322b81e528539b86efc90355e131b784f39797e1172aad361bd990b1de105  tests/visual/tooltip.html
67d8c2fbd86b0e18739b809dab8f1d1af9cfbf7f3bc2cd96e2507df5cb6e03cb  tooltip.js
1fd0bac6d1f9c7c8105290fb4e260eb4e35fcdd581128db9f090ce611715c0c6  transition.js
```

Of course, you can also specify multiple directories at once.

[MD5]: https://en.wikipedia.org/wiki/MD5
[FNV]: https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
[SHA]: https://en.wikipedia.org/wiki/Secure_Hash_Algorithms
