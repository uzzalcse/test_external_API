package main

import (
	"fmt"

	"github.com/uzzalcse/test_external_api/services"
)

func main() {
	u := services.GetUser()
	for _, v := range u {
		fmt.Println(v)
	}
}