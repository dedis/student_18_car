// Status takes in a file containing a list of servers and returns the status
// reports of all of the servers.  A status is a list of connections and
// packets sent and received for each server in the file.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	status "github.com/dedis/cothority/status/service"
	"github.com/dedis/onet"
	"github.com/dedis/onet/app"
	"github.com/dedis/onet/log"
	"github.com/dedis/onet/network"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Status"
	app.Usage = "Get and print status of all servers of a file."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "group, g",
			Value: "group.toml",
			Usage: "Cothority group definition in `FILE.toml`",
		},
		cli.StringFlag{
			Name:  "format, f",
			Value: "txt",
			Usage: "Output format: \"txt\" (default) or \"json\".",
		},
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: `integer`: 1 for terse, 5 for maximal",
		},
	}
	app.Action = func(c *cli.Context) error {
		log.SetUseColors(false)
		log.SetDebugVisible(c.GlobalInt("debug"))
		return action(c)
	}
	app.Run(os.Args)
}

type se struct {
	Server *network.ServerIdentity
	Status *status.Response
	Err    string
}

// will contact all cothorities in the group-file and print
// the status-report of each one.
func action(c *cli.Context) error {
	groupToml := c.GlobalString("g")
	format := c.String("format")

	el, err := readGroup(groupToml)
	log.ErrFatal(err, "Couldn't Read File")
	log.Lvl3(el)
	cl := status.NewClient()

	var all []se
	for _, server := range el.List {
		sr, err := cl.Request(server)
		if err != nil {
			err = fmt.Errorf("could not get status from %v: %v", server, err)
		}

		if format == "txt" {
			if err != nil {
				log.Print(err)
			} else {
				printTxt(sr)
			}
		} else {
			// JSON
			errStr := "ok"
			if err != nil {
				errStr = err.Error()
			}
			all = append(all, se{Server: server, Status: sr, Err: errStr})
		}
	}
	if format == "json" {
		printJSON(all)
	}
	return nil
}

// readGroup takes a toml file name and reads the file, returning the entities within
func readGroup(tomlFileName string) (*onet.Roster, error) {
	f, err := os.Open(tomlFileName)
	if err != nil {
		return nil, err
	}
	g, err := app.ReadGroupDescToml(f)
	if err != nil {
		return nil, err
	}
	if len(g.Roster.List) <= 0 {
		return nil, errors.New("Empty or invalid group file:" +
			tomlFileName)
	}
	log.Lvl3(g.Roster)
	return g.Roster, err
}

// prints the status response that is returned from the server
func printTxt(e *status.Response) {
	var a []string
	if e.Status == nil {
		log.Print("no status from ", e.ServerIdentity)
		return
	}

	for sec, st := range e.Status {
		for key, value := range st.Field {
			a = append(a, (sec + "." + key + ": " + value))
		}
	}
	sort.Strings(a)
	log.Print(strings.Join(a, "\n"))
}

func printJSON(all []se) {
	b1 := new(bytes.Buffer)
	e := json.NewEncoder(b1)
	e.Encode(all)

	b2 := new(bytes.Buffer)
	json.Indent(b2, b1.Bytes(), "", "  ")

	out := bufio.NewWriter(os.Stdout)
	out.Write(b2.Bytes())
	out.Flush()
}
