package bml

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		b    []byte
		want *Elem
	}{
		{
			b: []byte("<a href=\"https://github.com/kybin/bml\">[bml repo]"),
			want: &Elem{
				children: []XMLMarshaler{
					&Elem{
						name:     "a",
						startTag: "<a href=\"https://github.com/kybin/bml\">",
						endTag:   "</a>",
						children: []XMLMarshaler{
							Text("bml repo"),
						},
					},
				},
			},
		},
		{
			b: []byte("<script> [``a < b``]"),
			want: &Elem{
				children: []XMLMarshaler{
					&Elem{
						name:     "script",
						startTag: "<script>",
						endTag:   "</script>",
						children: []XMLMarshaler{
							Text("a < b"),
						},
					},
				},
			},
		},
		{
			b: []byte("<a> [ hi! <b> [<c> [`main`]] bye! ]"),
			want: &Elem{
				children: []XMLMarshaler{
					&Elem{
						name:     "a",
						startTag: "<a>",
						endTag:   "</a>",
						children: []XMLMarshaler{
							Text(" hi! "),
							&Elem{
								name:     "b",
								startTag: "<b>",
								endTag:   "</b>",
								children: []XMLMarshaler{
									&Elem{
										name:     "c",
										startTag: "<c>",
										endTag:   "</c>",
										children: []XMLMarshaler{
											Text("main"),
										},
									},
								},
							},
							Text(" bye! "),
						},
					},
				},
			},
		},
		{
			b: []byte(`<script />
<script />`),
			want: &Elem{
				children: []XMLMarshaler{
					&Elem{
						name:     "script",
						startTag: "<script />",
						endTag:   "",
						children: nil,
					},
					Text("\n"),
					&Elem{
						name:     "script",
						startTag: "<script />",
						endTag:   "",
						children: nil,
					},
				},
			},
		},
	}
	for _, c := range cases {
		got, err := Unmarshal(c.b)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("want.children: %v", c.want.children)
		t.Logf("got.children: %v", got.children)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("Unmarshal(%s): want %s, got %s", c.b, c.want.XMLMarshal(), got.XMLMarshal())
		}
	}
}

func TestWithData(t *testing.T) {
	cases := []struct {
		bml string
		xml string
	}{
		{
			bml: "testdata/test.bml",
			xml: "testdata/test.html",
		},
	}
	for _, c := range cases {
		in, err := ioutil.ReadFile(c.bml)
		if err != nil {
			t.Fatal(err)
		}
		gotEl, err := Unmarshal(in)
		if err != nil {
			t.Fatal(err)
		}
		got := gotEl.XMLMarshal()
		if err != nil {
			t.Fatal(err)
		}
		want, err := ioutil.ReadFile(c.xml)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatalf("unexpected xml result: %s: want %s\n\n\ngot %s", c.bml, want, got)
		}
	}
}
