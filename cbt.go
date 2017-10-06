package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LoliGothick/cbt/internal/solution"
	"github.com/LoliGothick/cbt/internal/wandbox"
	"github.com/LoliGothick/cbt/internal/wandbox/expand"
	"github.com/LoliGothick/freyja/cutil"
	"github.com/LoliGothick/freyja/maybe"
	"github.com/LoliGothick/freyja/set"
	"github.com/urfave/cli"
)

type CLI struct {
	app *cli.App
}

func main() {
	NewCLI().Run()
}

func NewCLI() *CLI {
	_cli := new(CLI)

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
						cli.BoolFlag{
							Name: "bash",
						},
					},
				},
				{
					Name:   "c",
					Usage:  "C Language",
					Action: WandboxC,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "compiler, x",
							Value: "gcc-head",
						},
						cli.StringFlag{
							Name:  "std",
							Value: "c11",
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
						cli.BoolFlag{
							Name: "bash",
						},
					},
				},
			},
		},
		{
			Name:    "solution",
			Aliases: []string{"sln"},
			Usage:   "solution management",
			Subcommands: []cli.Command{
				{
					Name:   "init",
					Usage:  "Initialize solution",
					Action: SolutionInitial,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "lang",
							Value: "cpp",
						},
					},
				},
				{
					Name:   "update",
					Usage:  "update solution",
					Action: SolutionUpdate,
				},
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("(´･_･`)? Command not found")
		return nil
	}
	_cli.app = app
	return _cli
}

func (_cli *CLI) Run() {
	_cli.app.Run(os.Args)
}

func (_cli *CLI) TestRun(args []string) ([]byte, error) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	_cli.app.Writer = outStream
	_cli.app.ErrWriter = errStream
	err := _cli.app.Run(args)
	return outStream.Bytes(), err
}

func WandboxC(c *cli.Context) {
	// preprocessing

	// prepare JSON struct
	config := wandbox.Request{}
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
	if !c.Bool("bash") {
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
		options += c.String("x")
		if len(c.Args()) < 2 {
			code, codes := expand.ExpandInclude(string(c.Args().First()), `#include.*".*"|".*"/\*cbt-require\*/`)
			// JSON configure
			config = wandbox.Request{
				Compiler:          c.String("x") + "-c",
				Code:              code,
				Codes:             wandbox.TransformToCodes(codes),
				Options:           options,
				Stdin:             string(stdin),
				CompilerOptionRaw: c.String("c"),
				RuntimeOptionRaw:  c.String("r"),
				Save:              c.Bool("s"),
			}
		} else {
			targets := []string{}
			targets = c.Args()
			code, src, codes := expand.ExpandIncludeMulti(targets, `#include.*".*"|".*"/\*cbt-require\*/`)
			config = wandbox.Request{
				Compiler:          c.String("x") + "-c",
				Code:              code,
				Codes:             wandbox.TransformToCodes(codes),
				Options:           options,
				Stdin:             string(stdin),
				CompilerOptionRaw: strings.Join(src, "\n") + "\n" + c.String("c"),
				RuntimeOptionRaw:  c.String("r"),
				Save:              c.Bool("s"),
			}
		}
	} else {
		{ // else target is multiple src-file
			// set target
			target := c.Args()
			// code analyze
			codes := expand.ExpandAll(target, `#include.*".*"|".*"/\*cbt-require\*/`)
			// generate template (shell)
			shell_tmpl := `
echo 'compiler:' {{.Compiler}}
echo 'target:' {{.Target}}
{{if .Clang}}
/opt/wandbox/{{.Compiler}}/bin/clang {{.Target}} {{.Option}} && ./a.out{{else}}/opt/wandbox/{{.Compiler}}/bin/g++ {{.Target}} -std={{.CXX}} {{.Option}} && ./a.out{{end}}{{if .StdinFlag}} <<- EOS
{{.Stdin}}
EOS{{end}}
`
			options := ""
			if c.Bool("w") {
				options += ` -Wall -Wextra`
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

			tmpl := template.Must(template.New("bash").Parse(shell_tmpl))
			bash := &wandbox.Bash{
				Compiler:  c.String("x") + "-c",
				Target:    strings.Join(target, " "),
				CXX:       "c",
				VER:       c.String("std")[1:],
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
				Codes:    wandbox.TransformToCodes(codes),
				Save:     c.Bool("s"),
			}
		}
	}
	postRequest(config, c.Bool("s"), c.App.Writer, c.App.ErrWriter)
}

func WandboxCpp(c *cli.Context) {
	// preprocessing

	// prepare JSON struct
	config := wandbox.Request{}
	// prepare stdin
	var stdin string
	switch in := cutil.OrElse(c.String("in") == "", "", maybe.Expected(ioutil.ReadFile(c.String("in"))).UnwrapOr(c.String("in"))); in.(type) {
	case []byte:
		stdin = string(in.([]byte))
	case string:
		stdin = in.(string)
	case error:
		panic(in.(error))
	}

	// Let's Making JSON!
	if !c.Bool("bash") {
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
		if len(c.Args()) < 2 {
			code, codes := expand.ExpandInclude(string(c.Args().First()), `#include.*".*"|".*"/\*cbt-require\*/`)
			// JSON configure
			config = wandbox.Request{
				Compiler:          c.String("x"),
				Code:              code,
				Codes:             wandbox.TransformToCodes(codes),
				Options:           options,
				Stdin:             string(stdin),
				CompilerOptionRaw: c.String("c"),
				RuntimeOptionRaw:  c.String("r"),
				Save:              c.Bool("s"),
			}
		} else {
			targets := []string{}
			targets = c.Args()
			code, src, codes := expand.ExpandIncludeMulti(targets, `#include.*".*"|".*"/\*cbt-require\*/`)
			config = wandbox.Request{
				Compiler:          c.String("x"),
				Code:              code,
				Codes:             wandbox.TransformToCodes(codes),
				Options:           options,
				Stdin:             string(stdin),
				CompilerOptionRaw: strings.Join(src, "\n") + "\n" + c.String("c"),
				RuntimeOptionRaw:  c.String("r"),
				Save:              c.Bool("s"),
			}
		}
	} else {
		{ // else target is multiple src-file
			// set target
			target := c.Args()
			// code analyze
			codes := expand.ExpandAll(target, `#include.*".*"|".*"/\*cbt-require\*/`)
			// generate template (shell)
			shell_tmpl := `
echo 'compiler:' {{.Compiler}}
echo 'target:' {{.Target}}
{{if .Clang}}
/opt/wandbox/{{.Compiler}}/bin/clang++ {{.Target}} {{.Option}} && ./a.out{{else}}/opt/wandbox/{{.Compiler}}/bin/g++ {{.Target}} -std={{.CXX}}++{{.VER}} {{.Option}} && ./a.out{{end}}{{if .StdinFlag}} <<- EOS
{{.Stdin}}
EOS{{end}}
`
			cxx := strings.Split(c.String("std"), "++")
			options := ""
			if c.Bool("w") {
				options += ` -Wall -Wextra`
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
				CXX:       cxx[0],
				VER:       cxx[1],
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
				Codes:    wandbox.TransformToCodes(codes),
				Save:     c.Bool("s"),
			}
		}
	}
	postRequest(config, c.Bool("s"), c.App.Writer, c.App.ErrWriter)
}

func postRequest(config wandbox.Request, save bool, stdout, stderr io.Writer) bool {
	// Marshal JSON
	cppJSONBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	json.Indent(out, cppJSONBytes, "", "    ") // pretty

	file, err := os.Create(`./config.json`)
	if err != nil {
		panic(err)
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
		stdout.Write([]byte(result.ProgramMessage))
	case result.CompilerError != "":
		stdout.Write([]byte("Compilation Error!:"))
		stdout.Write([]byte(result.CompilerError))
	case result.ProgramError != "":
		stdout.Write([]byte("Runtime Error!:"))
		stdout.Write([]byte(result.ProgramError))
	}

	if save {
		stdout.Write([]byte("Permlink: " + result.Permlink))
		stdout.Write([]byte("URL: " + result.URL))
	}
	return true
}

func SolutionInitial(c *cli.Context) {
	name := c.Args().First()

	if name == "" {
		fmt.Println(`empty solution name`)
		return
	}

	if err := os.MkdirAll(name+"/"+name, 0777); err != nil {
		fmt.Println(err)
	}

	sln := solution.Sln{
		Name: name,
		Lang: c.String("lang"),
		Project: []solution.Project{
			solution.Project{
				Name:   name,
				Type:   "Application",
				Target: []string{},
				Module: []string{},
			},
		},
	}
	// Marshal JSON
	slnJSONBytes, err := json.Marshal(sln)
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	json.Indent(out, slnJSONBytes, "", "    ") // pretty

	prev, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	defer os.Chdir(prev)

	os.Chdir(name)

	file, err := os.Create(name + `.cbt.json`)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	file.Write(([]byte)(out.String()))
}

func SolutionUpdate(c *cli.Context) {
	info := new(solution.Info)
	sol, _ := filepath.Glob("*.cbt.json")

	if len(sol) == 0 {
		fmt.Println(`solution not found.`)
		return
	}

	if len(sol) != 1 {
		panic(`2 or more solution files found!`)
	}

	sln := new(solution.Sln)

	b, err := ioutil.ReadFile(sol[0])
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, sln); err != nil {
		panic(err)
	}

	var update []solution.Project

	switch sln.Lang {
	case "cpp":
		for _, project := range sln.Project {
			target := set.StringSet{}
			module := set.StringSet{}
			for _, path := range project.Target {
				if cutil.FileExists(path) {
					target.Add(path)
					info.Add()
				} else {
					fmt.Println(`delete target`, path)
					info.Delete()
				}
			}
			for _, path := range project.Module {
				if cutil.FileExists(path) {
					module.Add(path)
				} else {
					fmt.Println(`delete module`, path)
					info.Delete()
				}
			}
			project.Target = []string{}
			project.Module = []string{}
			filepath.Walk(`./`+project.Name+`/`, func(path string, finfo os.FileInfo, err error) error {
				if filepath.Ext(path) == ".cpp" {
					fmt.Println(`add target ` + path)
					target.Add(path)
					info.Add()
				}
				if filepath.Ext(path) == ".hpp" {
					fmt.Println(`add modules ` + path)
					module.Add(path)
					info.Add()
				}
				return nil
			})
			target.Range(func(p string) interface{} {
				project.Target = append(project.Target, p)
				return nil
			})
			module.Range(func(p string) interface{} {
				project.Module = append(project.Module, p)
				return nil
			})
			update = append(update, solution.Project{
				Name:   project.Name,
				Type:   project.Type,
				Target: project.Target,
				Module: project.Module,
			})
		}
	default:

	}

	sln.Project = update

	if jb, err := json.Marshal(sln); err != nil {
		panic(err)
	} else {
		out := new(bytes.Buffer)
		json.Indent(out, jb, "", "    ") // pretty
		fmt.Println(out)
		file, err := os.Create(sol[0])
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		file.Write(([]byte)(out.String()))
	}
	info.Show()
}
