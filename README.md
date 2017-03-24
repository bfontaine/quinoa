# Quinoa

## Build

    $ go build -o quinoa ./main.go

## Run

    # code
    $ touch foo.qi

    # compile
    $ ./quinoa -o foo ./foo.qi

    # run
    $ ./foo

The parsing step is not implemented so it always generates the same object code
for now.

## Hacking

1. Install LLVM Go bindings using [GoCaml’s script][goscript]:

        git clone --depth=1 https://github.com/rhysd/gocaml && cd gocaml/scripts
        ./install_llvmgo.sh

[goscript]: https://github.com/g/blob/master/scripts/install_llvmgo.sh

## References

* [An introduction to LLVM in Go](https://felixangell.com/blog/an-introduction-to-llvm-in-go)
