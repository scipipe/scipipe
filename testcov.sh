#!/bin/bash
set -ex
echo "mode: count" > cover.out
for pkg in "github.com/scipipe/scipipe" "github.com/scipipe/scipipe/components" "github.com/scipipe/scipipe/cmd/scipipe"; do
    if [[ -f profile_tmp.cov ]]; then
        rm profile_tmp.cov;
    fi
    touch profile_tmp.cov
    go test -v -covermode=count -coverprofile=profile_tmp.cov $pkg || ERROR="Error testing $pkg"
    tail -n +2 profile_tmp.cov >> cover.out || exit "Unable to append coverage for $pkg"
done
