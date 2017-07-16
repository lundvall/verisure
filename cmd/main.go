package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lundvall/verisure"
)

func must(cancel context.CancelFunc, err error) {
	if err != nil {
		cancel()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}

func main() {
	username := flag.String("username", "", "Verisure username")
	password := flag.String("password", "", "Verisure password")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client := verisure.New()

	must(cancel, client.Login(ctx, *username, *password))

	o, err := client.Overview(ctx)
	must(cancel, err)
	fmt.Println(o)

	must(cancel, client.Logout(ctx))
}
