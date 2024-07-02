#!/bin/bash

target=service_types_data.json

if [ -z "$UPCLOUD_USERNAME" ] || [ -z "$UPCLOUD_PASSWORD" ]; then
    echo "Error: UpCloud credentials not set."
    exit 1
fi

cat $target | jq > old_$target

# Fetch DB types data and remove extra fields and whitespace. End-result is a single line JSON file with DB properties objects by database type.
upctl database types -o json | jq -cM 'map_values({properties})' > $target

cat $target | jq > new_$target

if diff -u old_$target new_$target > $target.diff; then
    echo "No changes detected."
else
    echo "Changes detected. Please review the diff file: $target.diff"
    echo ""
    cat $target.diff
fi

rm -f old_$target new_$target
