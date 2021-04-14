package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/MaxHayter/IS_BAYES_HW2/internal/bayes"
	"github.com/MaxHayter/IS_BAYES_HW2/internal/importing"
)

const file = "credit_net.json"

func main() {
	f := flag.String("f", file, "import file")
	flag.Parse()

	gr, err := importing.ImportGraph(*f)
	if err != nil {
		log.Fatalln(err)
	}

	b := bayes.NewBayes(gr)

	var command string
	var exit bool
	var splitCommand []string
	var numCol int

	reader := bufio.NewReader(os.Stdin)

	for !exit {
		fmt.Println("Enter command:")
		command, err = reader.ReadString('\n')
		if err != nil {
			log.Fatalln(err)
		}
		splitCommand = strings.Split(command[:len(command)-1], " ")
		if len(splitCommand) == 0 {
			log.Println("Введите корректную команду")
			continue
		}

		switch splitCommand[0] {
		case "help":
			fmt.Println("Доступные команды:")
			fmt.Println("print <NUM_COL> - Вывод графа")
			fmt.Println("set NAME_NODE STATE - Сделать вершину наблюдаемой")
			fmt.Println("unset NAME_NODE - Сделать вершину ненаблюдаемой")
			fmt.Println("exit - Выйти")
		case "print":
			numCol = 4
			if len(splitCommand) > 1 {
				if num, err := strconv.Atoi(splitCommand[1]); err == nil {
					numCol = num
				}
			}
			b.Graph.PrettyPrint(numCol)
		case "set":
			if len(splitCommand) < 3 {
				log.Println("Неправильный формат команды")
				continue
			}
			err = b.SetConstant(splitCommand[1], splitCommand[2])
			if err != nil {
				log.Println(err)
				continue
			}
		case "unset":
			if len(splitCommand) < 2 {
				log.Println("Неправильный формат команды")
				continue
			}
			err = b.UnsetConstant(splitCommand[1])
			if err != nil {
				log.Println(err)
				continue
			}
		case "exit":
			exit = true
		}
		fmt.Println("Done!")
	}
}
