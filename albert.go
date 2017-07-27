// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package goalbert implements the Albert Communication Protocol
*/
package goalbert

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

// Albert Communication Protocol version we support
const protocolVersion = "org.albert.extension.external/v2.0"

var pluginMetadata = Metadata{IID: protocolVersion}
var defaultHookSet = HookSet{
	Metadata:        defaultMetadata,
	Name:            func() int { fmt.Print(pluginMetadata.Name); return 0 },
	Initialize:      func() int { return 0 },
	Finalize:        func() int { return 0 },
	SetupSession:    func() int { return 0 },
	TeardownSession: func() int { return 0 },
	Query: func(query string) (result QueryResult, code int) {
		glog.Warning("Must implement Query behavior")
		return QueryResult{}, 255
	},
}

// HookSet represents a set of callbacks to run in different stages of the
// interaction between Albert and the plugin
type HookSet struct {
	Metadata        func() (code int)
	Name            func() (code int)
	Initialize      func() (code int)
	Finalize        func() (code int)
	SetupSession    func() (code int)
	TeardownSession func() (code int)
	Query           func(query string) (result QueryResult, code int)
}

// Start initiates the Albert Communication Protocol
func (h *HookSet) Start() {
	switch op := os.Getenv("ALBERT_OP"); op {
	case "METADATA":
		{
			if fn := h.Metadata; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.Metadata())
			}
		}
	case "NAME":
		{
			if fn := h.Name; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.Name())
			}
		}
	case "INITIALIZE":
		{
			if fn := h.Initialize; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.Initialize())
			}
		}
	case "FINALIZE":
		{
			if fn := h.Finalize; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.Finalize())
			}
		}
	case "SETUPSESSION":
		{
			if fn := h.SetupSession; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.SetupSession())
			}
		}
	case "TEARDOWNSESSION":
		{
			if fn := h.TeardownSession; fn != nil {
				os.Exit(fn())
			} else {
				os.Exit(defaultHookSet.TeardownSession())
			}
		}
	case "QUERY":
		{
			query := os.Getenv("ALBERT_QUERY")
			if fn := h.Query; fn != nil {
				handleQueryResult(fn(query))
			} else {
				handleQueryResult(defaultHookSet.Query(query))
			}

		}
	}
}

// QueryResult is the collection of QueryItem that is output to Albert every
// time it asks for new query
type QueryResult struct {
	Items []QueryItem `json:"items"`
}

// QueryItem represents an item in the list of items to show to a user while
// they're typing their query. Each item can have multiple actions to take
type QueryItem struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Completion  string        `json:"completion"`
	Icon        string        `json:"icon"`
	Actions     []QueryAction `json:"actions"`
}

// QueryAction represents an action to take when a user presses Return on a
// given query
type QueryAction struct {
	Name      string   `json:"name"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}

// Metadata represents the information used to describe your plugin to Albert
type Metadata struct {
	IID          string   `json:"iid"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Author       string   `json:"author"`
	Dependencies []string `json:"dependencies"`
	Trigger      string   `json:"trigger"`
}

// NewQueryAction returns an action given a name and exec.Cmd
func NewQueryAction(name string, cmd *exec.Cmd) QueryAction {
	return QueryAction{Name: name, Command: cmd.Path, Arguments: cmd.Args[1:]}
}

// SetInfo can be used to set basic information about your plugin
func SetInfo(name, version, author string) {
	pluginMetadata.Name = name
	pluginMetadata.Version = version
	pluginMetadata.Author = author
}

// SetTrigger sets the trigger for users to type into Albert to signal desire
// to use your plugin
func SetTrigger(trigger string) {
	pluginMetadata.Trigger = trigger
}

// SetDependencies sets the dependencies needed to run your plugin
func SetDependencies(deps []string) {
	pluginMetadata.Dependencies = deps
}

// SetMetadata can be used to directly set the plugin metadata from
// pre-constructed struct if that is preferred over using SetInfo, SetTrigger,
// and SetDependencies
func SetMetadata(m *Metadata) {
	pluginMetadata = *m
}

func defaultMetadata() int {
	js, _ := json.Marshal(pluginMetadata)
	fmt.Print(string(js))
	return 0
}

func handleQueryResult(res QueryResult, code int) {
	js, err := json.Marshal(res)
	if err != nil {
		glog.Warning(err)
		os.Exit(255)
	}
	fmt.Print(string(js))
	os.Exit(code)
}
