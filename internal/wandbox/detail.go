// Package wandbox is cbt internal package.
// Analyze code and create JSON for wandbox API.
package wandbox

// Bash struct for shell script(bash) text/template
type Bash struct {
	Compiler  string
	Target    string
	CXX       string
	VER       string
	Option    string
	StdinFlag bool
	Stdin     string
	Clang     bool
}

// Code is JSON Object for WandboxRequest
type Code struct {
	FileName string `json:"file"`
	Code     string `json:"code"`
}

// Request is JSON struct
type Request struct {
	Compiler          string `json:"compiler"`
	Code              string `json:"code"`
	Codes             []Code `json:"codes,omitempty"`
	Options           string `json:"options,omitempty"`
	Stdin             string `json:"stdin,omitempty"`
	CompilerOptionRaw string `json:"compiler-option-raw,omitempty"`
	RuntimeOptionRaw  string `json:"runtime-option-raw,omitempty"`
	Save              bool   `json:"save,omitempty"`
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

func TransformToCodes(m map[string]string) []Code {
	var ret []Code
	for name, code := range m {
		ret = append(ret, Code{name, code})
	}
	return ret
}

func TransformToMap(codes []Code) map[string]string {
	ret := map[string]string{}
	for _, data := range codes {
		ret[data.FileName] = data.Code
	}
	return ret
}
