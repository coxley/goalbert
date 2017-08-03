package goalbert

import "os/exec"

func ExampleNewQueryAction() {
	cmd := exec.Command("google-chrome-stable", "www.facebook.com")
	action := NewQueryAction("Browse to Facebook", cmd)
}

func ExampleQueryResult() {
	toFb := NewQueryAction("Browse to Facebook", exec.Command("google-chrome-stable", "www.facebook.com"))
	toGoog := NewQueryAction("Browse to Google", exec.Command("google-chrome-stable", "www.google.com"))

	QueryResult{
		Items: []QueryItem{
			QueryItem{"1", "Visit Facebook", "", "", []QueryAction{toFb}},
		},
		Items: []QueryItem{
			QueryItem{"1", "Visit Google", "", "", []QueryAction{toGoog}},
		},
		Items: []QueryItem{
			QueryItem{"1", "Visit Facebook and Google", "", "", []QueryAction{toFb, toGoog}},
		},
	}
}
