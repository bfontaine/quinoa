# Quinoa

## Run

For now it’s not that interesting:

    $ go build ./main.go
    $ ./main

It should generate `a.o`. Link it with `ld a.o -lSystem -o foo` (on macOS).
Then:

    $ ./foo
    $ echo $?
    42

That’s it!

## Hacking

1. Install LLVM Go bindings using [GoCaml’s script][goscript]:

        git clone --depth=1 https://github.com/rhysd/gocaml && cd gocaml/scripts
        ./install_llvmgo.sh

[goscript]: https://github.com/g/blob/master/scripts/install_llvmgo.sh
