package mario

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func ConfigSetup() {

}

func parseInput(input string) string {
	fmt.Println("input text:")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return line
}
