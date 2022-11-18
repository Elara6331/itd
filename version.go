package main

import _ "embed"

//go:generate scripts/gen-version.sh

//go:embed version.txt
var version string
