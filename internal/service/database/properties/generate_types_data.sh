#!/bin/bash

tmpfile=tmp_typesdata.json

if [ -z "$UPCLOUD_USERNAME" ] || [ -z "$UPCLOUD_PASSWORD" ]; then
    echo "Error: UpCloud credentials not set."
    exit 1
fi

upctl database types -o json | jq -cM > $tmpfile
types=$(cat $tmpfile | jq -r 'keys | .[]')

for type in $types; do
    cat $tmpfile | jq .$type.properties -cM > ${type}_properties.json
done

rm -f $tmpfile
