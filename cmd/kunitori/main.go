package kunitori

import (
	"io/ioutil"
	"log"
)

func main() {
	log.SetOutput(ioutil.Discard)
	println("moro")
}
