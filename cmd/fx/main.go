package main

import (
	"fmt"
	"go.uber.org/fx"
)

type Data string
type Data2 string

type Node struct {
	fx.In

	Data Data //`name:"data"`
}

func main() {
	node := &Node{}
	fx.New(
		fx.Populate(node),
		fx.Provide(func(dd Data2) Data {
			return Data(dd) + "_sss"
		}),
		fx.Supply(Data2("test")),
		//fx.Options(
		//	fx.Module("data",
		//
		//	)),
	)
	fmt.Printf("node: %+3v", node)
}
