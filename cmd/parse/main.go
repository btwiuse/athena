package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirkon/goproxy/gomod"
)

func main() {
	modfile := os.Args[1]
	bytes, err := ioutil.ReadFile(modfile)
	if err != nil {
		panic(err)
	}

	mod, err := gomod.Parse(modfile, bytes)
	if err != nil {
		panic(err)
	}

	fmt.Println(mod.Name)
}
