package iniparse

import (
	"fmt"
	"io/ioutil"
)

// Section represents a section of a ini file
// headed by [Name]
type Section struct {
	Name string
	Keys map[string]string
}

// KeyIsPresent returns true if the key exists in the section
// and false if not
func (s *Section) KeyIsPresent(key string) bool {
	if _, ok := s.Keys[key]; ok {
		return true
	}
	return false
}

// String converts an Section to human readable output
func (s *Section) String() string {
	res := fmt.Sprintf("[%s]\n", s.Name)
	for k, v := range s.Keys {
		res += fmt.Sprintf("%s=%s\n", k, v)
	}
	return res
}

// IniFile represents an ini file
type IniFile struct {
	Path     string
	Sections []*Section
}

// NewIniFile constructs an IniFile from a file path
func NewIniFile(path string) (*IniFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error: could not read file: %s\n", path)
	}
	tokens := tokenize(data)
	sections := parse(tokens)
	return &IniFile{path, sections}, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parse(tokens []string) []*Section {
	var sections []*Section
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
			section := &Section{Name: header, Keys: keys}
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
