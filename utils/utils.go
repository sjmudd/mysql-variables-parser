package utils

import (
	"fmt"

	"code.google.com/p/go.net/html"
)

// PrintToken prints a formatted token
func PrintToken(token html.Token) {
	fmt.Println("tokenType:", token.Type, ":", token.Data)
	for i := range token.Attr {
		fmt.Println(" Attribute:", i, "=", token.Attr[i])
	}
}
