// Package expand is cbt internal package.
// Analyze C++ code and create JSON for wandbox API.
package expand

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func unique(path string, m map[string]string) string {
	for _, ok := m[path]; ok; _, ok = m[path] {
		path = "_" + path
	}
	return path
}

func flat(rm [][]string) []string {
	ret := []string{}
	for _, s := range rm {
		str := strings.Join(s, "")
		ret = append(ret, str[strings.Index(str, "\"")+1:strings.LastIndex(str, "\"")])
	}
	return ret
}

type StringSlice []string

func (ss StringSlice) Split(pred func(string) bool) (string, []string) {
	var main = ""
	sub := []string{}

	for _, target := range ss {
		switch {
		case pred(target):
			main = target
		default:
			sub = append(sub, filepath.Base(target))
		}
	}
	return main, sub
}

// ExpandInclude : Expand only included files(for one file compilation)
func ExpandInclude(file string, re string) (string, map[string]string) {
	return Expand([]string{file}, file, re)
}

// ExpandIncludeMulti : Expand all files(for muliple file compilation)
func ExpandIncludeMulti(files StringSlice, re string) (string, []string, map[string]string) {
	mre := regexp.MustCompile(`main\([\s\S]*?\){[\s\S]*?}`)
	main, sub := files.Split(func(target string) bool {
		src, err := ioutil.ReadFile(target)
		if err != nil {
			panic(err)
		}
		return mre.Match(src)
	})
	prog, headers := ExpandMulti(files, main, sub, re)
	return prog, sub, headers
}

// ExpandAll : Expand all files(for muliple file compilation)
func ExpandAll(files []string, re string) map[string]string {
	_, ret := Expand(files, "false", re)
	return ret
}

func Expand(files []string, src string, re string) (string, map[string]string) {
	init := map[string]string{}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		init[file] = abs
	}
	object := analyzingTo(init, re)

	prog := ""
	if src != "false" {
		abs, err := filepath.Abs(src)
		if err != nil {
			panic(err)
		}
		prog = object[abs]
		delete(object, abs)
	} else {
		for _, cpp := range files {
			abs, err := filepath.Abs(cpp)
			if err != nil {
				panic(err)
			}
			tmp := object[abs]
			delete(object, abs)
			object[filepath.Base(abs)] = tmp
		}
	}
	return prog, object
}

func ExpandMulti(files []string, src string, sub []string, re string) (string, map[string]string) {
	init := map[string]string{}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		init[file] = abs
	}
	object := analyzingTo(init, re)
	abs, err := filepath.Abs(src)
	if err != nil {
		panic(err)
	}
	prog := object[abs]
	delete(object, abs)
	for _, del := range sub {
		abs, err := filepath.Abs(del)
		if err != nil {
			panic(err)
		}
		tmp := object[abs]
		delete(object, abs)
		object[filepath.Base(abs)] = tmp
	}
	return prog, object
}

func analyzingTo(files map[string]string, re string) map[string]string {
	regex := regexp.MustCompile(re)
	rest := map[string]string{}
	mapCodes := map[string]string{}
	mapPath := map[string]string{}
	target := files
BACKTRACING:
	for file, rename := range target {
		mapCodes[rename] = ""
		dir := filepath.Dir(file)
		src, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		matched := regex.FindAllStringSubmatch(string(src), -1)
		if len(matched) == 0 {
			mapCodes[rename] = string(src)
			continue
		} else {
			for _, include := range matched {
				str := strings.Join(include, "")
				path := str[strings.Index(str, "\"")+1 : strings.LastIndex(str, "\"")]
				next := filepath.Join(dir, path)
				absNext, err := filepath.Abs(next)
				if err != nil {
					panic(err)
				}
				if _, ok := mapPath[absNext]; ok {
					src = regexp.MustCompile(regexp.QuoteMeta(path)).ReplaceAll(src, ([]byte)(mapPath[absNext]))
					continue
				} else {
					xtRename := unique(filepath.Base(path), mapCodes)
					mapPath[absNext] = xtRename
					rest[next] = xtRename
					src = regexp.MustCompile(regexp.QuoteMeta(path+`"`)).ReplaceAll(src, ([]byte)(xtRename+`" /* origin >>> `+strings.Join(include, "")+" */"))
				}
			}
			mapCodes[rename] = string(src)
		}
	}
	if len(rest) != 0 {
		target = rest
		rest = make(map[string]string)
		goto BACKTRACING
	}

	return mapCodes
}