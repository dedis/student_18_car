package main

import (
	// Service needs to be imported here to be instantiated.
	_ "github.com/dedis/student_18_car/service"
	"github.com/dedis/onet/simul"
)

func main() {
	simul.Start()
}
