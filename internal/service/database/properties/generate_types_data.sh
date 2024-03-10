#!/bin/bash

target=service_types_data.json

if [ -z "$UPCLOUD_USERNAME" ] || [ -z "$UPCLOUD_PASSWORD" ]; then
    echo "Error: UpCloud credentials not set."
    exit 1
fi

upctl database types -o json | jq -cM 'map_values({properties})' > $target
