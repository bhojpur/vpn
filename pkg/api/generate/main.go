package main

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func main() {
	templateFile := os.Args[1]
	src := os.Args[2]
	output := os.Args[3]

	b, err := ioutil.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}
	b2, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}

	templated, err := TemplatedString(fmt.Sprintf("%s\n%s", string(b), string(b2)), nil)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(output, []byte(templated), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func TemplatedString(t string, i interface{}) (string, error) {
	b := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(t)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(b, i)

	return b.String(), err
}
