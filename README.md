[![Go](https://github.com/carbonin/overreader/actions/workflows/go.yml/badge.svg)](https://github.com/carbonin/overreader/actions/workflows/go.yml)

# overreader

Implements a single reader which handles replacing multiple ranges in an underlying reader with separate data.

## Usage

```go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/carbonin/overreader"
)

func main() {
	sr := strings.NewReader("12345 ----- abcde ===")

	range1 := &overreader.Range{
		Content: strings.NewReader("67890"),
		Offset:  6,
	}
	range2 := &overreader.Range{
		Content: strings.NewReader("xyz"),
		Offset:  18,
	}

	r, err := overreader.NewReader(sr, range1, range2)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(os.Stdout, r); err != nil {
		log.Fatal(err)
	}
	fmt.Println("")
}
```

Running the program above results in the output: `12345 67890 abcde xyz`
