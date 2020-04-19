# bml
bml is a xml/html variant with brackets.

currently Unmarshal doesn't check the original xml is well-formed.

these are equivalent.

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

if you need to treat content of a tag as raw string
add backquote(`) to brackets.

to avoid collision with content, number of backquotes is your choice
but ensure those are matched at both open, close brackets.

a tag with backquotes is always a leaf tag.

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
