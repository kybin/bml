# bml
bml is a xml/html variant with brackets.

I created it because I wanted to read/write html templates without use of xml syntax.
(which is hard to debug in my editor.)

So it has a replacement of html/template.ParseGlob function.

```
tmpl := template.Must(bml.ToHTMLParseGlob("", template.FuncMap{}, "tmpl/*.bml"))
```


### Note

Unmarshal doesn't check the original xml is well-formed.
So it can decode html as well.


### Syntax

These are equivalent.

bml

```
<a> [
	<b> []
	<c>
]
```

xml/html
```
<a>
	<b></b>
	<c>
</a>
```

#### Tags with backquotes

If you need to treat content of a tag as raw string
add backquote(`) to brackets.

To avoid collision with the content, number of backquotes is your choice,
but ensure those are matched at both open, close brackets.

A tag with backquotes is always a leaf tag.

bml
```
<script> [``
function min(a, b) {
	if (a < b) {
		return a
	}
	return b
}
``]
```
