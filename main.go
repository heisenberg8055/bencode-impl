package main

import (
	"fmt"
	"strings"

	bencode "github.com/heisenberg8055/bencode-impl/internal"
)

func main() {
	test := strings.NewReader("i02e")
	reff, err := bencode.Decode(test)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(reff)
}
