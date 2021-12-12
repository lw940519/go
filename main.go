package main

import (
	"ceshi/week2/service"
	"fmt"
)


func main() {
	_, err := service.GetUserList()
	if err != nil {
		fmt.Printf("user not found, %v\n", err)
		fmt.Printf("%+v\n", err)
	}
}