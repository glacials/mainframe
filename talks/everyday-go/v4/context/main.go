package main

import (
	"context"
)

func main() {
	ctx := context.Background()
	ctx.Value("user-id")
}
