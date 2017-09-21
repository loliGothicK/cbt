package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/LoliGothick/cbt/internal/wandbox"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "cbt (Cranberries Build Tool)"
	app.Usage = "C++ Build Tool"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		// command config
		{
			Name:    "wandbox",
			Aliases: []string{"wb"},
			Usage:   "Send codes to wandbox.org",
			Action:  Wandbox,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "compiler, x",
					Value: "gcc-head",
				},
				cli.StringFlag{
					Name:  "std",
					Value: "c++14",
				},
				cli.StringFlag{
					Name:  "stdin,in",
					Value: "",
				},
				cli.StringFlag{
					Name:  "option, o",
					Value: "warnning",
				},
				cli.StringFlag{
					Name:  "compile-option, c",
					Value: "",
				},
				cli.StringFlag{
					Name:  "runtime-option, r",
					Value: "",
				},
				cli.BoolFlag{
					Name: "save, s",
				},
			},
		},
	}

	// app.Before = func(c *cli.Context) error {
	// 	fmt.Println("Build Start. plz wait...")
	// 	return nil
	// }

	// app.After = func(c *cli.Context) error {
	// 	fmt.Println("Successfuly!")
	// 	return nil
	// }

	app.Action = func(c *cli.Context) error {
		fmt.Println("(´･_･`)")
		return nil
	}

	app.Run(os.Args)
}

func Wandbox(c *cli.Context) {
	// preprocessing

	// prepare JSON struct
	config := wandbox.Request{}
	// regex for C++
	analyzer := wandbox.Analyzer{Regex: regexp.MustCompile(`#include.*".*"|".*"/\*Option\*/`)}
	// prepare stdin
	stdin := ([]byte)("")
	if c.String("in") != "" {
		_, err := os.Stat(c.String("in"))
		if err == nil {
			stdin, err = ioutil.ReadFile(c.String("in"))
			if err != nil {
				panic(err)
			}
		}
	}

	// Let's Making JSON!
	if len(c.Args()) < 2 { // if target is a src-file
		code, codes := analyzer.ExpandInclude(string(c.Args().First()))
		// JSON configure
		config = wandbox.Request{
			Compiler:          c.String("x"),
			Code:              code,
			Codes:             codes,
			Options:           strings.Join([]string{c.String("o"), c.String("std")}, ","),
			Stdin:             string(stdin),
			CompilerOptionRaw: c.String("c"),
			RuntimeOptionRaw:  c.String("r"),
			Save:              c.Bool("s"),
		}
	} else { // else target is multiple src-file
		// set target
		target := c.Args()
		// code analyze
		codes := analyzer.ExpandAll(target)
		// generate template (shell)
		tmpl := template.Must(template.ParseFiles("bash.tmpl"))
		bash := &wandbox.Bash{
			c.String("c"),
			strings.Join(target, " "),
			strings.Join([]string{c.String("o"), c.String("std"), c.String("c")}, " "),
			string(stdin) != "false",
			string(stdin),
			string(("clang-head")[0:3]) != "gcc",
		}
		var shell = ""
		buf := bytes.NewBufferString(shell)
		err := tmpl.Execute(buf, bash)
		if err != nil {
			panic(err)
		}
		// JSON configure
		config = wandbox.Request{
			Compiler: "bash",
			Code:     buf.String(),
			Codes:    codes,
			Save:     c.Bool("s"),
		}
	}

	// Marshal JSON
	cppJSONBytes, err := json.Marshal(config)
	if err != nil {
		fmt.Println("JSON config marshal error:", err)
		return
	}
	out := new(bytes.Buffer)
	json.Indent(out, cppJSONBytes, "", "    ") // pretty

	file, err := os.Create(`./config.json`)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.Write(([]byte)(out.String()))

	// Client : Wait Time 30s
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	// Request : POST JSON
	req, err := http.NewRequest("POST", "https://wandbox.org/api/compile.json?", strings.NewReader(out.String()))
	if err != nil {
		panic(err)
	}
	// Header : Content-type <- application/json
	req.Header.Add("Content-type", "application/json")

	// Send POST
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	result := new(wandbox.Result)
	if err := json.Unmarshal(([]byte)(body), result); err != nil {
		panic(err)
	}
	fmt.Println(result.ProgramMessage)
}
