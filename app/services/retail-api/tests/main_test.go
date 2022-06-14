package tests

import (
	"fmt"
	"testing"

	"github.com/dimashiro/service/business/data/tests"
	"github.com/dimashiro/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = tests.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tests.StopDB(c)

	m.Run()
}
