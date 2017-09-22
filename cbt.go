package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
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
		{
			Name:    "wandbox",
			Aliases: []string{"wb"},
			Usage:   "Send your codes to wandbox",
			Subcommands: []cli.Command{
				{
					Name:   "cpp",
					Usage:  "C++",
					Action: WandboxCpp,
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
						cli.BoolFlag{
							Name: "warning, w",
						},
						cli.StringFlag{
							Name:  "pedantic, p",
							Value: "no",
						},
						cli.BoolFlag{
							Name: "verbose, v",
						},
						cli.BoolFlag{
							Name: "optimize, o",
						},
						cli.BoolFlag{
							Name: "sprout",
						},
						cli.StringFlag{
							Name:  "boost",
							Value: "nothing",
						},
						cli.BoolFlag{
							Name: "msgpack, m",
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
		fmt.Println("(´･_･`)? Command not found")
		return nil
	}

	app.Run(os.Args)
}

func WandboxCpp(c *cli.Context) {
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
		options := c.String("std")
		if c.Bool("w") {
			options += ",warning"
		}
		switch c.String("p") {
		case "no":
			options += ",cpp-no-pedantic"
		case "yes":
			options += ",cpp-pedantic"
		case "errors":
			options += ",cpp-pedantic-errors"
		}
		if c.Bool("v") {
			options += ",cpp-verbose"
		}
		if c.Bool("o") {
			options += ",optimize"
		}
		if c.Bool("sprout") {
			options += ",sprout"
		}
		if c.Bool("msgpack") {
			options += ",msgpack"
		}
		options += ",boost-" + c.String("boost") + "-" + c.String("x")

		// JSON configure
		config = wandbox.Request{
			Compiler:          c.String("x"),
			Code:              code,
			Codes:             codes,
			Options:           options,
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
		shell_tmpl := `
echo 'compiler:' {{.Compiler}}
echo 'target:' {{.Target}}
{{if .Clang}}
/opt/wandbox/{{.Compiler}}/bin/clang++ {{.Target}} {{.Option}} && ./a.out{{else}}/opt/wandbox/{{.Compiler}}/bin/g++ {{.Target}} {{.Option}} && ./a.out{{end}}{{if .StdinFlag}} <<- EOS
{{.Stdin}}
EOS{{end}}
`
		options := "-std=" + c.String("std")
		if c.Bool("w") {
			options += " -Wall -Wextra"
		}
		switch c.String("p") {
		case "no":
		case "yes":
			options += " -pedantic"
		case "errors":
			options += " -pedantic-errors"
		}
		if c.Bool("v") {
			options += " -v"
		}
		if c.Bool("o") {
			options += " -O2 -march=native"
		}
		if c.Bool("sprout") {
			options += " -I/opt/wandbox/sprout"
		}
		if c.Bool("msgpack") {
			options += " -I/opt/wandbox/msgpack/include"
		}
		if c.String("boost") != "nothing" {
			options += " -I/opt/wandbox/boost-" + c.String("boost") + "/" + c.String("x") + "/include"
		}

		tmpl := template.Must(template.New("bash").Parse(shell_tmpl))
		bash := &wandbox.Bash{
			Compiler:  c.String("x"),
			Target:    strings.Join(target, " "),
			Option:    options,
			StdinFlag: string(stdin) != "",
			Stdin:     string(stdin),
			Clang:     c.String("x")[0:3] != "gcc",
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

	switch {
	case result.ProgramMessage != "":
		fmt.Println(result.ProgramMessage)
	case result.CompilerError != "":
		fmt.Println("Compilation Error!:")
		fmt.Println(result.CompilerError)
	case result.ProgramError != "":
		fmt.Println("Runtime Error!:")
		fmt.Println(result.ProgramError)
	}

	if c.Bool("s") {
		fmt.Println("Permlink:", result.Permlink)
		fmt.Println("URL:", result.URL)
	}
}
