#!/bin/bash
while read oldrev newrev refname
do
    branch=$(git rev-parse --symbolic --abbrev-ref $refname)
    if [ "master" = "$branch" ]; then
        echo "wrong branch"
        exit 1
    fi
done