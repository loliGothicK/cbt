// Package wandbox is cbt internal package.
// Analyze code and create JSON for wandbox API.
package wandbox

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// Bash struct for shell script(bash) text/template
type Bash struct {
	Compiler  string
	Target    string
	Option    string
	StdinFlag bool
	Stdin     string
	Clang     bool
}

// AdditionalCode is JSON Object for WandboxRequest
type AdditionalCode struct {
	FileName string `json:"file"`
	Code     string `json:"code"`
}

// Request is JSON struct
type Request struct {
	Compiler          string           `json:"compiler"`
	Code              string           `json:"code"`
	Codes             []AdditionalCode `json:"codes,omitempty"`
	Options           string           `json:"options,omitempty"`
	Stdin             string           `json:"stdin,omitempty"`
	CompilerOptionRaw string           `json:"compiler-option-raw,omitempty"`
	RuntimeOptionRaw  string           `json:"runtime-option-raw,omitempty"`
	Save              bool             `json:"save,omitempty"`
}

// Result is JSON struct
type Result struct {
	Status          string `json:"status"`
	Signal          string `json:"signal"`
	CompilerOutput  string `json:"compiler_output"`
	CompilerError   string `json:"compiler_error"`
	CompilerMessage string `json:"compiler_messagestdin"`
	ProgramOutput   string `json:"program_output"`
	ProgramError    string `json:"program_error"`
	ProgramMessage  string `json:"program_message"`
	Permlink        string `json:"permlink"`
	URL             string `json:"url"`
}

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

// Analyzer is Code Analyzer.
// Expanding Include codes.
type Analyzer struct {
	Regex *regexp.Regexp
	Codes map[string]string
	Path  map[string]string
}

// ExpandInclude : Expand only included files(for one file compilation)
func (analyzer *Analyzer) ExpandInclude(file string) (string, []AdditionalCode) {
	return analyzer.expand([]string{file}, file)
}

// ExpandAll : Expand all files(for muliple file compilation)
func (analyzer *Analyzer) ExpandAll(files []string) []AdditionalCode {
	_, ret := analyzer.expand(files, "false")
	return ret
}

func (analyzer *Analyzer) expand(files []string, src string) (string, []AdditionalCode) {
	if analyzer.Codes == nil {
		analyzer.Codes = map[string]string{}
	}
	if analyzer.Path == nil {
		analyzer.Path = map[string]string{}
	}

	init := map[string]string{}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			panic(err)
		}
		init[file] = abs
	}
	ret := []AdditionalCode{}
	object := analyzer.analyzingTo(init)
	prog := "false"
	if src != "false" {
		abs, err := filepath.Abs(src)
		if err != nil {
			panic(err)
		}
		prog = object[abs]
		delete(object, abs)
	} else {
		for _, file := range files {
			abs, err := filepath.Abs(file)
			if err != nil {
				panic(err)
			}
			tmp := object[abs]
			delete(object, abs)
			object[filepath.Base(abs)] = tmp
		}
	}
	for file, code := range object {
		ret = append(ret, AdditionalCode{file, code})
	}
	return prog, ret
}

func (analyzer *Analyzer) analyzingTo(files map[string]string) map[string]string {
	var rest = map[string]string{}

	for file, rename := range files {
		analyzer.Codes[rename] = ""
		dir := filepath.Dir(file)
		src, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		matched := analyzer.Regex.FindAllStringSubmatch(string(src), -1)
		if len(matched) == 0 {
			analyzer.Codes[rename] = string(src)
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
				if _, ok := analyzer.Path[absNext]; ok {
					src = regexp.MustCompile(regexp.QuoteMeta(path)).ReplaceAll(src, ([]byte)(analyzer.Path[absNext]))
					continue
				} else {
					xtRename := unique(filepath.Base(path), analyzer.Codes)
					analyzer.Path[absNext] = xtRename
					rest[next] = xtRename
					src = regexp.MustCompile(regexp.QuoteMeta(path)).ReplaceAll(src, ([]byte)(xtRename))
				}
			}
			analyzer.Codes[rename] = string(src)
		}

	}
	if len(rest) != 0 {
		return analyzer.analyzingTo(rest)
	}

	return analyzer.Codes
}
