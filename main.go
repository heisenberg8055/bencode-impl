package main

import (
	"fmt"
	"strings"

	bencode "github.com/heisenberg8055/bencode-impl/internal"
)

func main() {
	test := strings.NewReader("d7:meaning7:bencode4:wikii42ee")
	reff, err := bencode.Decode(test)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(reff)
}
