// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// https://github.com/casbin/casbin/blob/master/config/config.go
var (
	// DEFAULT_SECTION specifies the name of a section if no name provided
	DEFAULT_SECTION = "default"
	// DEFAULT_COMMENT defines what character(s) indicate a comment `#`
	DEFAULT_COMMENT = []byte{'#'}
	// DEFAULT_COMMENT_SEM defines what alternate character(s) indicate a comment `;`
	DEFAULT_COMMENT_SEM = []byte{';'}
	// DEFAULT_MULTI_LINE_SEPARATOR defines what character indicates a multi-line content
	DEFAULT_MULTI_LINE_SEPARATOR = []byte{'\\'}
)

type Reader interface {
	RequestDef() map[string]string
	PolicyDef() map[string]string
	PolicyEffect() map[string]string
	RoleDef() map[string]string
	Matchers() map[string]string
}

type config struct {
	data map[string]map[string]string
}

func (c *config) RoleDef() map[string]string {
	return c.data["role_definition"]
}

func (c *config) RequestDef() map[string]string {
	return c.data["request_definition"]
}

func (c *config) PolicyDef() map[string]string {
	return c.data["policy_definition"]
}

func (c *config) PolicyEffect() map[string]string {
	return c.data["policy_effect"]
}

func (c *config) Matchers() map[string]string {
	return c.data["matchers"]
}

// AddConfig adds a new section->key:value to the configuration.
func (c *config) AddConfig(section string, option string, value string) bool {
	if section == "" {
		section = DEFAULT_SECTION
	}

	if _, ok := c.data[section]; !ok {
		c.data[section] = make(map[string]string)
	}

	_, ok := c.data[section][option]
	c.data[section][option] = value

	return !ok
}

func (c *config) write(section string, lineNum int, b *bytes.Buffer) error {
	if b.Len() <= 0 {
		return nil
	}

	optionVal := bytes.SplitN(b.Bytes(), []byte{'='}, 2)
	if len(optionVal) != 2 {
		return fmt.Errorf("parse the content error : line %d , %s = ? ", lineNum, optionVal[0])
	}
	option := bytes.TrimSpace(optionVal[0])
	value := bytes.TrimSpace(optionVal[1])
	c.AddConfig(section, string(option), string(value))

	// flush buffer after adding
	b.Reset()

	return nil
}

func (c *config) parse(buf *bufio.Reader) error {
	var section string
	var lineNum int
	var buffer bytes.Buffer
	var canWrite bool
	for {
		if canWrite {
			if err := c.write(section, lineNum, &buffer); err != nil {
				return err
			} else {
				canWrite = false
			}
		}
		lineNum++
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			// force write when buffer is not flushed yet
			if buffer.Len() > 0 {
				if err := c.write(section, lineNum, &buffer); err != nil {
					return err
				}
			}
			break
		} else if err != nil {
			return err
		}

		line = bytes.TrimSpace(line)
		switch {
		case bytes.Equal(line, []byte{}), bytes.HasPrefix(line, DEFAULT_COMMENT_SEM),
			bytes.HasPrefix(line, DEFAULT_COMMENT):
			canWrite = true
			continue
		case bytes.HasPrefix(line, []byte{'['}) && bytes.HasSuffix(line, []byte{']'}):
			// force write when buffer is not flushed yet
			if buffer.Len() > 0 {
				if err := c.write(section, lineNum, &buffer); err != nil {
					return err
				}
				canWrite = false
			}
			section = string(line[1 : len(line)-1])
		default:
			var p []byte
			if bytes.HasSuffix(line, DEFAULT_MULTI_LINE_SEPARATOR) {
				p = bytes.TrimSpace(line[:len(line)-1])
				p = append(p, " "...)
			} else {
				p = line
				canWrite = true
			}

			end := len(p)
			for i, value := range p {
				if value == DEFAULT_COMMENT[0] || value == DEFAULT_COMMENT_SEM[0] {
					end = i
					break
				}
			}
			if _, err := buffer.Write(p[:end]); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewParse(buf *bufio.Reader) (Reader, error) {
	c := &config{make(map[string]map[string]string)}
	err := c.parse(buf)
	if err != nil {
		return nil, err
	}
	return c, nil
}
