#!/bin/sh -xe

for d in resources data-sources; do
    files=$(ls "docs/$d")

    for f in $files; do
        subcategory=$(jq -r ".\"$d\".\"$f\"" <subcategories.json)

        if [ "$subcategory" != "null" ]; then
            sed -i "s/^subcategory:.*/subcategory: $subcategory/" docs/$d/$f;
        fi;
    done;
done;
