#!/bin/bash

target=typesdata.gen.go
tmpfile=tmp_typesdata.json

if [ -z "$UPCLOUD_USERNAME" ] || [ -z "$UPCLOUD_PASSWORD" ]; then
    echo "Error: UpCloud credentials not set."
    exit 1
fi

upctl database types -o json | jq -cM > $tmpfile
types=$(cat $tmpfile | jq -r 'keys | .[]')

output=""
for type in $types; do
    output="${output}const ${type}PropertiesJSON = \`$(cat $tmpfile | jq .$type.properties -cM | sed 's/`/` + \"`\" + `/g')\`
"
done

cat << EOF > $target
//go:build !codeanalysis
// +build !codeanalysis

//nolint:all // Automatically generated with ./generate_types_data.sh
package properties

$output
EOF

rm -f $tmpfile