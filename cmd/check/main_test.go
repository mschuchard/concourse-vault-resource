package main

import (
	"os"
	_ "testing"
)

func ExampleMain() {
	// defer stdin close and establish workdir from argsp[1]
	defer os.Stdin.Close()

	// deliver test pipeline file content as stdin to "in" the same as actual pipeline execution
	os.Stdin, _ = os.OpenFile("fixtures/token_kv.json", os.O_RDONLY, 0o644)

	main()
	// Output: [{"version":"1"}]
}
