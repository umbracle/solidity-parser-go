#!/bin/bash

# Create the grammar
dir=tree-sitter-src
mkdir -p $dir

cd $dir && tree-sitter generate ../tree-sitter-sol/grammar.js
cd ..

# Create go bindings
godir=tree-sitter
mkdir -p $godir/tree_sitter

cp $dir/src/parser.c $godir
cp $dir/src/tree_sitter/parser.h $godir/tree_sitter/

rm -rf $dir
