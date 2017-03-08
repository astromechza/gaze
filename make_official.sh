#!/usr/bin/env bash

set -e

VERSION_NUM=$(cat VERSION)

function buildbinary {
    goos=$1
    goarch=$2

    echo "Building official $goos $goarch release"

    name="gaze-${VERSION_NUM}_${goos}_${goarch}"
    outputfolder="build/$name"
    echo "Output Folder $outputfolder"
    mkdir -pv $outputfolder

    export GOOS=$goos
    export GOARCH=$goarch

    govvv build -i -v -o "$outputfolder/gaze" github.com/AstromechZA/gaze

    tar -czvf "build/$name.tar.gz" -C "build" "$name"
    ls -lh "build/$name.tar.gz"
    echo
}

rm -rfv build/

# build local
unset GOOS
unset GOARCH
govvv build -i -v github.com/AstromechZA/gaze

# build for mac
buildbinary darwin amd64

# build for linux
buildbinary linux amd64
