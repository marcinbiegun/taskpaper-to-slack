package main

import (
	"strconv"
	"testing"
)

func TestLineDepth(t *testing.T) {
	result := lineDepth("something 			")
	if result != 0 {
		t.Error("Wrong line depth: " + strconv.Itoa(result))
	}

	result = lineDepth("	- task	")
	if result != 1 {
		t.Error("Wrong line depth: " + strconv.Itoa(result))
	}
}

func TestIsTodayHeader(t *testing.T) {
	result := isTodayHeader("Header: @something()")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}

	result = isTodayHeader("		Header: @slack(asd)")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}
	result = isTodayHeader("		Header: @slack(123,asd)")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}

	result = isTodayHeader("		Header: @slack(2019-01-27,ab1c50e)")
	if result != true {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}
}

func TestExtractToday(t *testing.T) {
	content := `
asdasd
	- asdasdasd
Header with no tag:
Greater thing:
	Today tasks: @slack(messageid)
		- do something
			nested text
		- and other stuff
		- nested task
		some text
		Header inside: 
			- task
other shit
`

	target := `	Today tasks: @slack(messageid)
		- do something
			nested text
		- and other stuff
		- nested task
		some text
		Header inside: 
			- task`

	result := extractToday(content)
	if result != target {
		t.Error("Not parsed correctly, got:\n" + result)
	}
}
