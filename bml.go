package bml

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Elem struct {
	name     string
	startTag string
	children []XMLMarshaler
	endTag   string // might be empty
}

func (e *Elem) XMLMarshal() []byte {
	var b bytes.Buffer
	b.WriteString(e.startTag)
	for _, c := range e.children {
		b.Write(c.XMLMarshal())
	}
	b.WriteString(e.endTag)
	return b.Bytes()
}

type Text []byte

func (t Text) XMLMarshal() []byte {
	return []byte(t)
}

func (t Text) String() string {
	return string(t)
}

type XMLMarshaler interface {
	XMLMarshal() []byte
}

func Unmarshal(b []byte) (*Elem, error) {
	xms, _, err := unmarshal(b, 0, "", nil)
	return &Elem{children: xms}, err
}

func unmarshal(b []byte, i int, closeBracket string, parent *Elem) ([]XMLMarshaler, int, error) {
	if i == len(b) {
		return nil, i, nil
	}
	xms := make([]XMLMarshaler, 0)
	for i < len(b) {
		// expect a close bracket or a child.
		j := bytes.Index(b[i:], []byte("<"))
		if j < 0 {
			j = len(b[i:])
		}
		closing := false
		if closeBracket != "" {
			// closeBracket will be empty string, when it's root
			k := bytes.Index(b[i:], []byte(closeBracket))
			if k < 0 {
				k = len(b[i:])
			}
			if k < j {
				j = k
				closing = true
			}
		}
		if j == len(b[i:]) {
			if parent != nil {
				return nil, i, fmt.Errorf("doesn't end properly: %d", i)
			}
			// read of the entire bytes is done.
			xms = append(xms, Text(b[i:]))
			i = len(b)
			return xms, i, nil
		}
		if j != 0 {
			// if k := bytes.Index(b[i:i+j], []byte("[")); k >= 0 {
			// 	return nil, k, fmt.Errorf("unexpected open bracket in the content of bml. use &#91; intead")
			// }
			xms = append(xms, Text(b[i:i+j]))
		}
		i += j
		if closing {
			// return to the parent
			return xms, i, nil
		}
		// b[i] == "<", get a new elem.
		e := &Elem{}
		j = bytes.Index(b[i:], []byte(">"))
		if j < 0 {
			return nil, i, fmt.Errorf("tag does not end")
		}
		if j == 0 {
			return nil, i, fmt.Errorf("empty tag is invalid")
		}
		e.startTag = string(b[i : i+j+1])
		e.name = strings.Fields(e.startTag[1 : len(e.startTag)-1])[0]
		i += j + 1
		if i >= len(b) {
			break
		}
		// tag end. maybe followed by bracket.
		nspace := len(b[i:]) - len(bytes.TrimLeft(b[i:], " \t\n"))
		if nspace != 0 {
			// does not count space from > to [.
			i += nspace
		}
		if i >= len(b) {
			break
		}
		if b[i] == '[' {
			i += 1
			// find number of tailing backquotes
			// when a bracket starts with [``, it will find ``] close bracket.
			nbq := len(b[i:]) - len(bytes.TrimLeft(b[i:], "`"))
			i += nbq
			closeb := strings.Repeat("`", nbq) + "]"
			var ch []XMLMarshaler
			if nbq != 0 {
				// raw string tag
				j := bytes.Index(b[i:], []byte(closeb))
				if j == len(b[i:]) {
					return nil, i, fmt.Errorf("raw string tag not ended")
				}
				ch = []XMLMarshaler{Text(b[i : i+j])}
				i += j
			} else {
				// normal tag
				var err error
				ch, i, err = unmarshal(b, i, closeb, e)
				if err != nil {
					return nil, i, err
				}
				if string(b[i:i+len(closeb)]) != closeb {
					return nil, i, fmt.Errorf("content does not properly parsed")
				}
			}
			e.children = ch
			e.endTag = "</" + e.name + ">"
			i += len(closeb)
		} else {
			e.endTag = ""
		}
		xms = append(xms, e)
		// find the next tag.
	}
	if i > len(b) {
		return nil, 0, fmt.Errorf("index out of range while parsing")
	}
	return xms, i, nil
}

func ToHTMLParseGlob(name string, fmap template.FuncMap, pattern string) (*template.Template, error) {
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("bml: pattern matches no files: %#q", pattern)
	}
	t := template.New(name).Funcs(fmap)
	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		el, err := Unmarshal(b)
		if err != nil {
			return nil, err
		}
		xb := el.XMLMarshal()
		s := string(xb)
		name := filepath.Base(filename)
		// If this file has the same name as t,
		// this file becomes the contents of t,
		// so `t, err := New(name).Funcs(xxx).ParseFiles(name)` works.
		// Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
