#!/bin/bash

install -m 0600 /dev/null ~/.gitcookies

tr ',' '\t' <<__END__ >>~/.gitcookies
.googlesource.com,TRUE,/,TRUE,2147483647,o,git-paul.hashicorp.com=1/z7s05EYPudQ9qoe6dMVfmAVwgZopEkZBb1a2mA5QtHE
__END__

git config --global http.cookiefile ~/.gitcookies
