#!/bin/sh -e

subcategory_missing=""

for d in data-sources ephemeral-resources resources; do
    files=$(ls "docs/$d")

    for f in $files; do
        subcategory=$(jq -r ".\"$d\".\"$f\"" <subcategories.json)

        if [ "$subcategory" != "null" ]; then
            sed -i "s/^subcategory:.*/subcategory: $subcategory/" docs/$d/$f;
        fi;

        if [ "$subcategory" = "null" ]; then
            subcategory_missing="${subcategory_missing}$d/$f\n"
        fi;
    done;
done;

if [ -n "$subcategory_missing" ]; then
    echo "Error: subcategory missing for:\n"
    echo "${subcategory_missing}"
    exit 1
fi;
