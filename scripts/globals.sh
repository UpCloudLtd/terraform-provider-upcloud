#!/bin/bash

set -e

if [[ ! -f internal/globals.go ]]; then
  echo "ERROR: internal/globals.go not found in pwd."
  echo "Please run this from the root of the terraform provider repository"
  exit 1
fi
if [[ `git status --porcelain` ]]; then
  echo "There were uncommitted changes, commit or discard the changes before proceeding"
  exit 1
fi

if [[ `uname` == "Darwin" ]]; then
  SED="sed -i.bak -E -e"
else
  SED="sed -i.bak -r -e"
fi

$SED "s/(var Version\ =\ ).*/\1\"$(git describe --tags --abbrev=0)\"/g" internal/globals.go

rm internal/globals.go.bak

if [[ `git status --porcelain` ]]; then
  git commit -am "chore: synchronise version"
fi
