package main

import (
	"bufio"
	"fmt"
	menu "goProjects/gokedex/Menu"
	pokecache "goProjects/gokedex/pokeCache"
	"os"
	"strings"
)

const Prompt string = "Gokedex >:"

// infinite loop waiting for input
func main() {

	reader := bufio.NewReader(os.Stdin)

	menu.DisplayMenu()
	cache := pokecache.NewCache()
	menu.Setup(cache)

	for {
		//print prompt
		fmt.Println(Prompt)
		//input command
		input, err := reader.ReadString('\n')
		//break for args
		parts := strings.Fields(input)
		//fmt.Println(parts)
		//fmt.Println(parts[0])

		if len(parts) <= 0 {
			fmt.Println("Must input a command")
		}

		if len(parts) > 0 {
			firstWord := parts[0]
			firstChar := strings.ToLower(string(firstWord[0]))
			fmt.Println("Entry:", firstChar)

			if len(parts) > 1 {
				fmt.Println("Input:", parts[1])
			}

			cmdkey := firstChar

			if err != nil {
				fmt.Println(err)
				continue
			}

			//TODO: this bullshit down here

			//_, ok := menu.CmdMap[cmdkey]
			//if !ok {
			//	fmt.Println("command not avail")
			//	continue
			//}

			for _, cmd := range menu.CmdMap {
				if cmdkey == cmd.KeyMap {
					fmt.Println(cmd.Desc)
					cmd.CallBack(parts[1:]...)
					break
				}
			}
		}
	}
}
