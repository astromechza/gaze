#!/usr/bin/env bash

set -e

# first build the version string
VERSION_NUM=0.4

# add the git commit id and date
VERSION="$VERSION_NUM (commit $(git rev-parse --short HEAD) @ $(git log -1 --date=short --pretty=format:%cd))"

function buildbinary {
    goos=$1
    goarch=$2

    echo "Building official $goos $goarch binary for version '$VERSION'"

    outputfolder="build/${goos}_${goarch}"
    echo "Output Folder $outputfolder"
    mkdir -pv $outputfolder

    export GOOS=$goos
    export GOARCH=$goarch

    go build -i -v -o "$outputfolder/gaze" -ldflags "-X \"main.GazeVersion=$VERSION\"" github.com/AstromechZA/gaze

    echo "Done"
    ls -lh "$outputfolder/gaze"
    file "$outputfolder/gaze"
    echo
}

# build local 
unset GOOS
unset GOARCH
go build -ldflags "-X \"main.GazeVersion=$VERSION\"" github.com/AstromechZA/gaze

# build for mac
buildbinary darwin amd64

# build for linux
buildbinary linux amd64

# zip up
tar -czf gaze-${VERSION_NUM}.tgz -C build .
ls -lh gaze-${VERSION_NUM}.tgz
file gaze-${VERSION_NUM}.tgz
