package qfx

type Token interface {
	Whitespace | Tag | Text
}

type Whitespace struct {
	Char  rune
	Count int
}

type Tag struct {
	Name string
}

type Text string
