// Analyze C++ code and create JSON for wandbox API.
package expand

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/LoliGothick/freyja/maybe"
)

func unique(path string, m map[string]string) string {
	for _, ok := m[path]; ok; _, ok = m[path] {
		path = "_" + path
	}
	return path
}

type PathSlice []string

func (ss PathSlice) Split(pred func(string) bool) (string, []string) {
	var main = ""
	sub := []string{}

	for _, target := range ss {
		switch {
		case pred(target):
			if abs, err := filepath.Abs(target); err != nil {
				panic(err)
			} else {
				main = abs
			}
		default:
			if abs, err := filepath.Abs(target); err != nil {
				panic(err)
			} else {
				sub = append(sub, abs)
			}
		}
	}
	return main, sub
}

func (ss PathSlice) ToAbs() PathSlice {
	ret := []string{}
	for _, path := range ss {
		abs, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		ret = append(ret, abs)
	}
	ss = ret
	return ss
}

func (ss PathSlice) ToBase() PathSlice {
	ret := []string{}
	for _, path := range ss {
		ret = append(ret, filepath.Base(path))
	}
	ss = ret
	return ss
}

// ExpandInclude : Expand only included files(for one file compilation)
func ExpandInclude(file string, re string) (string, map[string]string) {
	return Expand([]string{file}, file, re)
}

func ExpandRubyRequire(file string, re string) (string, map[string]string) {
	return ExpandRuby([]string{file}, file, re)
}

// ExpandIncludeMulti : Expand all files(for muliple file compilation)
func ExpandIncludeMulti(files PathSlice, re string) (string, []string, map[string]string) {
	mre := regexp.MustCompile(`main\([\s\S]*?\){[\s\S]*?}`)
	var main string
	var sub PathSlice
	main, sub = files.ToAbs().Split(func(target string) bool {
		src, err := ioutil.ReadFile(target)
		if err != nil {
			panic(err)
		}
		return mre.Match(src)
	})
	prog, headers := ExpandMulti(files.ToAbs(), main, sub, re)
	return prog, sub.ToBase(), headers
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
		init[file] = file
	}
	object := analyzingTo(init, re)
	prog := object[src]
	delete(object, src)
	for _, del := range sub {
		tmp := object[del]
		delete(object, del)
		object[filepath.Base(del)] = tmp
	}
	return prog, object
}

func ExpandRuby(files []string, src string, re string) (string, map[string]string) {
	init := map[string]string{}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		init[file] = abs
	}
	object := analyzingToRuby(init, re)
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

func analyzingTo(files map[string]string, re string) map[string]string {
	regex := regexp.MustCompile(re)
	rest := map[string]string{}
	mapCodes := map[string]string{}
	mapPath := map[string]string{}
	target := files
BACKTRACING:
	for file, renamed := range target {
		mapCodes[renamed] = ""
		dir := filepath.Dir(file)
		src, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		matched := regex.FindAllStringSubmatch(string(src), -1)
		if len(matched) == 0 {
			mapCodes[renamed] = string(src)
			continue
		} else {
			pathRegex := regexp.MustCompile(`['|"](.*)['|"]`)
			for _, include := range matched {
				str := strings.Join(include, "")
				path := pathRegex.FindAllStringSubmatch(str, 1)[0][1]
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
					mapCodes[xtRename] = ""
					mapPath[absNext] = xtRename
					rest[next] = xtRename
					src = regexp.MustCompile(regexp.QuoteMeta(path+`"`)).ReplaceAll(src, ([]byte)(xtRename+`" /* origin >>> `+strings.Join(include, "")+" */"))
				}
			}
			mapCodes[renamed] = string(src)
		}
	}
	if len(rest) != 0 {
		target = rest
		rest = make(map[string]string)
		goto BACKTRACING
	}

	return mapCodes
}

func analyzingToRuby(files map[string]string, re string) map[string]string {
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
			pathRegex := regexp.MustCompile(`['|"](.*)['|"]`)
			for _, include := range matched {
				str := strings.Join(include, "")
				path := pathRegex.FindAllStringSubmatch(str, 1)[0][1] + ".rb"
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

func GOPATH() (*regexp.Regexp, *regexp.Regexp) {
	var ret []string
	var reg []string
	for _, path := range filepath.SplitList(build.Default.GOPATH) {
		switch dir := maybe.Expected(ioutil.ReadDir(path + `/src/`)).Interface(); dir.(type) {
		case []os.FileInfo:
			for _, m := range dir.([]os.FileInfo) {
				ret = append(ret, `"(`+m.Name()+`/.*)"`)
				reg = append(reg, regexp.QuoteMeta(path)+`(.*)`)
			}
		case error:
			panic(dir.(error))
		}
	}
	return regexp.MustCompile(strings.Join(ret, `|`)), regexp.MustCompile(strings.Join(reg, `|`))
}

func ReadDirEx(match string) ([]os.FileInfo, error) {

	for _, path := range filepath.SplitList(build.Default.GOPATH) {
		prevDir, _ := filepath.Abs(".")
		os.Chdir(path)
		files, err := ioutil.ReadDir(build.Default.GOPATH + `/src/` + match + `/`)
		os.Chdir(prevDir)
		if err == nil {
			return files, nil
		}
	}
	return nil, fmt.Errorf(`%s`, `path not found!`)
}

func ReadFileEx(match, name string) ([]byte, error) {

	for _, path := range filepath.SplitList(build.Default.GOPATH) {
		prevDir, _ := filepath.Abs(".")
		os.Chdir(path)
		files, err := ioutil.ReadFile(build.Default.GOPATH + `/src/` + match + `/` + name)
		os.Chdir(prevDir)
		if err == nil {
			return files, nil
		}
	}
	return nil, fmt.Errorf(`%s`, `path not found!`)
}

func ExpandGo(t string) (string, map[string]string) {
	regex, gopath := GOPATH()
	mapCodes := map[string]string{}
	target := []string{}
	rest := []string{}
	req := regexp.MustCompile("(\"|`)(.*)(\"|`).*" + regexp.QuoteMeta(`/*cbt-require*/`))

	switch b := maybe.Expected(ioutil.ReadFile(t)).Map(func(src []byte) bool {
		for _, opt := range req.FindAllStringSubmatch(string(src), -1) {
			mapCodes[opt[2]] = string(maybe.Expected(ioutil.ReadFile(opt[2])).Interface().([]byte))
		}
		for _, match := range regex.FindAllStringSubmatch(string(src), -1) {
			files, err := ReadDirEx(match[1])
			if err != nil {
				panic(err)
			}
			for _, f := range files {
				s, _ := ReadFileEx(match[1], f.Name())
				if regexp.MustCompile(`package ` + filepath.Base(match[1])).Match(s) {
					rest = append(rest, build.Default.GOPATH+`/src/`+match[1]+`/`+f.Name())
				}
			}
		}
		return true
	}).Interface(); b.(type) {
	case bool:
	case error:
		panic(b.(error))
	}
	by, _ := ioutil.ReadFile(t)
	main := (string)(by)

BACKTRACING:
	for _, pack := range target {
		switch b := maybe.Expected(ioutil.ReadFile(pack)).Map(func(src []byte) bool {
			// for _, opt := range req.FindAllStringSubmatch(string(src), -1) {
			// 	fmt.Println(opt[2])
			// 	mapCodes[opt[2]] = string(maybe.Expected(ioutil.ReadFile(opt[2])).Interface().([]byte))
			// }
			for _, match := range regex.FindAllStringSubmatch(string(src), -1) {
				files, err := ReadDirEx(match[1])
				if err != nil {
					panic(err)
				}
				for _, f := range files {
					s, _ := ReadFileEx(match[1], f.Name())
					if regexp.MustCompile(`package ` + filepath.Base(match[1])).Match(s) {
						rest = append(rest, build.Default.GOPATH+`/src/`+match[1]+`/`+f.Name())
					}
				}
			}

			mapCodes[string(gopath.ReplaceAll(([]byte)(pack), ([]byte)(`go$1`)))] = string(src)
			return true
		}).Interface(); b.(type) {
		case bool:
		case error:
			panic(b.(error))
		}
	}
	if len(rest) != 0 {
		target = rest
		rest = []string{}
		goto BACKTRACING
	}
	return main, mapCodes
}
