package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type Config struct {
	StructName      string
	PublicVariables bool
	Getters         bool
	Setters         bool
	Constructor     bool
	Package         string
}

func (cfg *Config) parseFlags() {
	flag.StringVar(&cfg.StructName, "name", "Foo", "Name for your struct")
	flag.BoolVar(&cfg.PublicVariables, "public", true, "make variables public")
	flag.BoolVar(&cfg.Getters, "getters", false, "make getters")
	flag.BoolVar(&cfg.Setters, "setters", false, "make setters")
	flag.BoolVar(&cfg.Constructor, "constructor", true, "make a constructor with empty arguments")
	flag.StringVar(&cfg.Package, "package", "main", "package name")

	flag.Parse()

	if cfg.Getters && cfg.PublicVariables {
		fmt.Fprintln(os.Stderr, "ERR: You can't have public variables and getters")
	}
}

func convertToCamelCase(input string, ucFirst bool) (output string) {
	parts := strings.Split(input, "_")

	for key, part := range parts {
		if len(part) >= 1 {
			if key == 0 && ucFirst == false {
				output += part
				continue
			}

			chars := []rune(part)
			chars[0] = unicode.ToUpper(chars[0])
			output += string(chars)
		}
	}

	return
}

func readStdin() []byte {
	tmp := make([]byte, 128)
	buf := make([]byte, 0, 2)
	reader := bufio.NewReader(os.Stdin)

	for {
		n, err := reader.Read(tmp)
		if nil != err {
			if err != io.EOF {
				fmt.Println("can't read from stdin")
				os.Exit(-1)
			}
			break
		}

		buf = append(buf, tmp[:n]...)
	}

	return buf
}

func typeOf(input interface{}) string {
	if val, ok := input.(float64); ok {
		if val == float64(int32(val)) {
			return "int"
		}

		if val == float64(int32(val)) {
			return "int64"
		}
		return "float64"
	}
	if _, ok := input.(string); ok {
		return "string"
	}
	return "interface{}"
}

func generateGetter(key string, val string, config *Config) (output string) {
	ccName := convertToCamelCase(key, true)
	ccKey := convertToCamelCase(key, config.PublicVariables)
	typeStr := typeOf(val)
	firstStr := string(unicode.ToLower([]rune(config.StructName)[0]))

	output += "func (" + firstStr + " *" + config.StructName + ") " + ccName + "() " + typeStr + " { \n"
	output += "\treturn " + firstStr + "." + ccKey + "\n"
	output += "}\n\n"

	return
}

func generateSetter(key string, val string, config *Config) (output string) {
	ccName := convertToCamelCase(key, true)
	ccKey := convertToCamelCase(key, config.PublicVariables)
	typeStr := typeOf(val)
	firstStr := string(unicode.ToLower([]rune(config.StructName)[0]))
	ccVarKey := convertToCamelCase(key, false)

	output += "func (" + firstStr + " *" + config.StructName + ") Set" + ccName + "(" + ccVarKey + " " + typeStr + ")  { \n"
	output += "\t" + firstStr + "." + ccKey + " = " + ccVarKey + "\n"
	output += "}\n\n"

	return
}

func generateConstructor(structName string) (output string) {
	output += "\nfunc New" + structName + "() *" + structName + " { \n"
	output += "\t o := new(" + structName + ")\n"
	output += "\t return o \n"
	output += "} \n\n"

	return
}

func main() {
	config := new(Config)
	config.parseFlags()

	content := readStdin()

	contentMap := make(map[string]interface{})

	err := json.Unmarshal(content, &contentMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}

	var output string

	output += "package " + config.Package + "\n\n"

	output += "type " + config.StructName + " struct { \n"
	for key, val := range contentMap {
		ccKey := convertToCamelCase(key, config.PublicVariables)
		output += "\t" + ccKey + " " + typeOf(val) + " `json:\"" + key + "\"` \n"
	}
	output += "}\n"

	if config.Constructor {
		output += generateConstructor(config.StructName)
	}

	for key, val := range contentMap {
		if config.Getters {
			output += generateGetter(key, val, config)
		}

		if config.Setters {
			output += generateSetter(key, val, config)
		}
	}

	fmt.Print(output)
}