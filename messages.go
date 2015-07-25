package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
)

type Format struct {
	Cmd    string      `json:"cmd,omitempty"`
	Status string      `json:"status,omitempty"`
	Err    string      `json:"error,omitempty"`
	Desc   string      `json:"desc,omitempty"`
	Uuid   []string    `json:"uuid,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func missingParameter(c *cli.Context, cmd string, param string, desc string) {
	fmtOutput(c,
		&Format{
			Cmd:  cmd,
			Err:  fmt.Sprintf("missing parameter %q", param),
			Desc: desc,
		})
}

func incorrectUsage(c *cli.Context, cmd string, desc string) {
	fmtOutput(c,
		&Format{
			Cmd:  cmd,
			Err:  fmt.Sprintf("incorrect usage"),
			Desc: desc,
		})
}

func badParameter(c *cli.Context, cmd string, param string, desc string) {
	fmtOutput(c,
		&Format{
			Cmd:  cmd,
			Err:  fmt.Sprintf("missing parameter %q", param),
			Desc: desc,
		})
}

func cmdError(c *cli.Context, cmd string, err error) {
	fmtOutput(c,
		&Format{
			Cmd: cmd,
			Err: err.Error(),
		})
}

func cmdErrorBody(c *cli.Context, cmd string, err error, body string) {
	body = strings.Replace(body, "\n", " ", -1)
	body = strings.Replace(body, "\"", "", -1)
	fmtOutput(c,
		&Format{
			Cmd:  cmd,
			Err:  err.Error(),
			Desc: body,
		})
}

func cmdOutput(c *cli.Context, body interface{}) {
	b, _ := json.MarshalIndent(body, "", " ")
	fmt.Printf("%+v\n", string(b))
}

func fmtOutput(c *cli.Context, format *Format) {
	jsonOut := c.GlobalBool("json")
	if jsonOut {
		b, _ := json.MarshalIndent(format, "", " ")
		fmt.Printf("%+v\n", string(b))
		return
	}
	if format.Err == "" {
		if format.Result == nil {
			for _, v := range format.Uuid {
				fmt.Println(v)
			}
			return
		}
		b, _ := json.MarshalIndent(format.Result, "", " ")
		fmt.Printf("%+v\n", string(b))
		return
	}
	if format.Desc != "" {
		fmt.Printf("%s: %v - %s\n", format.Cmd, format.Err, format.Desc)
		return
	}
	fmt.Printf("%s: %v\n", format.Cmd, format.Err)
}
