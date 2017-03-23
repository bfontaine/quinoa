# Quinoa

## Run

For now it’s not that interesting:

    $ go build ./main.go
    $ ./main

It should generate `a.out`.
Then:

    $ ./a.out
    $ echo $?
    42

That’s it! It runs on macOS; probably not on other platforms.

## Hacking

1. Install LLVM Go bindings using [GoCaml’s script][goscript]:

        git clone --depth=1 https://github.com/rhysd/gocaml && cd gocaml/scripts
        ./install_llvmgo.sh

[goscript]: https://github.com/g/blob/master/scripts/install_llvmgo.sh

## References

* [An introduction to LLVM in Go](https://felixangell.com/blog/an-introduction-to-llvm-in-go)
