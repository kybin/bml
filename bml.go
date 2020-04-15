package bml

import (
	"bytes"
	"fmt"
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
			xms = append(xms, Text(b[i:i+nspace]))
			i += nspace
		}
		if i >= len(b) {
			break
		}
		if b[i] == '[' {
			// find number of tailing backquotes
			// when a bracket starts with [``, it will find ``] close bracket.
			nbq := len(b[i+1:]) - len(bytes.TrimLeft(b[i+1:], "`"))
			closeb := strings.Repeat("`", nbq) + "]"
			ch, ii, err := unmarshal(b, i+nbq+1, closeb, e)
			if err != nil {
				return nil, ii, err
			}
			i = ii
			if string(b[i:i+len(closeb)]) != closeb {
				return nil, i, fmt.Errorf("content does not properly parsed")
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
