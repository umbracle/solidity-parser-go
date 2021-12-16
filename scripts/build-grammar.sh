#!/bin/bash

dir=./antlr
rm -rf $dir
mkdir $dir

# move grammar to $tmpDir
cp ./antlr-sol/Solidity.g4 $dir

# build the grammar
docker run --rm -u $(id -u ${USER}):$(id -g ${USER}) \
    -v `pwd`:/work ferranbt/antlr4 \
    -package solcparser \
    -Dlanguage=Go \
    -no-listener \
    -no-visitor \
    ${dir}/Solidity.g4
