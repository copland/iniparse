package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type section struct {
	name string
	keys map[string]string
}

func (p *section) String() string {
	res := fmt.Sprintf("[%s]\n", p.name)
	for k, v := range p.keys {
		res += fmt.Sprintf("%s=%s\n", k, v)
	}
	return res
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// TODO(copland): this is messy and should be cleaned up
// build an AST instead of iterating through the list
func parse(tokens []string) []*section {
	var sections []*section
	i := 0
	for i < len(tokens) {
		if tokens[i] == "[" {
			offset := i + 1
			header := tokens[offset]
			keys := make(map[string]string)

			for offset < len(tokens) && tokens[offset] != "[" {
				if tokens[offset] == "=" {
					keys[tokens[offset-1]] = tokens[offset+1]
				}
				offset++
			}
			section := &section{name: header, keys: keys}
			sections = append(sections, section)
			i = offset
		} else {
			i++
		}

	}
	return sections
}

func tokenize(stream []byte) []string {
	var tokens []string
	currToken := ""
	for i, streamCh := range stream {
		switch streamCh {
		case '[':
			tokens = append(tokens, string(streamCh))
			currToken = ""
		case ']', '=':
			tokens = append(tokens, currToken)
			currToken = ""
			tokens = append(tokens, string(streamCh))
		case '\n':
			if stream[i-1] != ']' && stream[i-1] != '\n' {
				tokens = append(tokens, currToken)
				currToken = ""
			}
		case ' ', '\t':
			continue
		default:
			currToken += string(streamCh)
		}
	}
	return tokens
}

func main() {
	if len(os.Args) < 1 {
		fmt.Printf("error: expected 1 argument: file path\n")
		os.Exit(1)
	}
	iniFile := os.Args[1]
	data, err := ioutil.ReadFile(iniFile)
	if err != nil {
		fmt.Printf("error: could not read file: %s\n", err)
		os.Exit(1)
	}

	tokens := tokenize(data)
	sections := parse(tokens)

	// TODO(copland): provide query capabilities

	for _, section := range sections {
		fmt.Printf("%s\n", section)
	}
}
