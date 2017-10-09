package solutions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/LoliGothick/cbt/internal/solution"
	"github.com/LoliGothick/freyja/cutil"
	"github.com/LoliGothick/freyja/set"
	"github.com/urfave/cli"
)

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
