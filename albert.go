// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package goalbert implements the Albert Communication Protocol

For full information about how Albert will interact with your program, see
the documentation here: https://albertlauncher.github.io/docs/extending/external/#communication-protocol-v2

Simplest example of setup:

	package main

	import (
		"os"
		"strings"
	)

	const name = "plugin_name"
	const version = "0.1"
	const trigger = "mytrigger"
	const author = "coxley"

	func query(q string) (albert.QueryResult, error) {
		// Strip the trigger if it's there from the query string
		q = strings.Replace(q, trigger+" ", "", 1)
		item := albert.QueryItem{
		// ...
		}
		return albert.QueryResult{Items: []albert.QueryItem{item}}, nil
	}

	func main() {
		p := albert.NewPlugin(name, version, author, trigger, query)

		// If called by Albert, run and exit based on Albert protocol
		if os.Getenv("ALBERT_OP") != "" {
			albert.Run(p)
		}
	}
*/
package goalbert

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

// Albert Communication Protocol version we support
const defaultProtocolVersion = "org.albert.extension.external/v2.0"

// AlbertOp is an operation requested by Albert at some point during the
// process
type AlbertOp string

// The operations as of defaultProtocolVersion
const (
	OpMetadata        AlbertOp = "METADATA"
	OpName            AlbertOp = "NAME"
	OpInitialize      AlbertOp = "INITIALIZE"
	OpFinalize        AlbertOp = "FINALIZE"
	OpSetupSession    AlbertOp = "SETUPSESSION"
	OpTeardownSession AlbertOp = "TEARDOWNSESSION"
	OpQuery           AlbertOp = "QUERY"
)

// AlbertError is an error thrown from an AlbertOp implementation containing
// both the error and exit-code to exit with
type AlbertError struct {
	Err  error
	Code int
}

func (e AlbertError) Error() string {
	return e.Err.Error()
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

// NewQueryAction returns an action given a name and exec.Cmd
func NewQueryAction(name string, cmd *exec.Cmd) QueryAction {
	return QueryAction{Name: name, Command: cmd.Path, Arguments: cmd.Args[1:]}
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

// Plugin is the primary interface for implementing Albert compatible plugins.
type Plugin interface {
	Metadata() Metadata
	Query(query string) (QueryResult, error)
	RunOp(op AlbertOp) error
}

// Run the given Plugin with the appropriate operation and exit accordingly
func Run(plugin Plugin) {
	op := AlbertOp(os.Getenv("ALBERT_OP"))
	err := plugin.RunOp(op)
	if err != nil {
		if aberr, ok := err.(AlbertError); ok {
			glog.Warningf("error (code=%d): %+v", aberr.Code, aberr)
			os.Exit(aberr.Code)
		}
		glog.Warningf("error: %+v", err)
		os.Exit(255)
	}
	os.Exit(0)
}

// DefaultPlugin is a default implementation of Plugin that can be embed for
// minimum rewriting sane behavior
//
// Meta describes the plugin
// Output is a Writer where JSON should be written to
type DefaultPlugin struct {
	Meta          Metadata
	Output        io.Writer
	QueryCallback func(query string) (QueryResult, error)
}

// NewPlugin configures a DefaultPlugin
func NewPlugin(name, version, author, trigger string, qc func(query string) (QueryResult, error)) DefaultPlugin {
	return DefaultPlugin{
		Meta: Metadata{
			IID:          defaultProtocolVersion,
			Name:         name,
			Version:      version,
			Author:       author,
			Trigger:      trigger,
			Dependencies: []string{},
		},
		Output:        os.Stdout,
		QueryCallback: qc,
	}
}

// Metadata returns a copy of the plugin's metadata
func (p DefaultPlugin) Metadata() Metadata {
	return p.Meta
}

// Query is one place where no sane default could exist. This must be
// implemented by you via the QueryCallback middleman
//
// query is a string that is input by the user into Albert and may inlude the
// trigger as part of the query
func (p DefaultPlugin) Query(query string) (QueryResult, error) {
	if p.QueryCallback != nil {
		return p.QueryCallback(query)
	}
	return QueryResult{}, AlbertError{
		Err:  fmt.Errorf("no behavior defined for query '%s'", query),
		Code: 255,
	}
}

// RunOp is to take any of operation and run, returning error if any, and
// writing to p.Output when appropriate
func (p DefaultPlugin) RunOp(op AlbertOp) error {
	switch op {
	case OpMetadata:
		err := json.NewEncoder(p.Output).Encode(p.Metadata())
		if err != nil {
			return err
		}
		return nil
	case OpName:
		_, err := p.Output.Write([]byte(p.Meta.Name))
		return err
	case OpInitialize:
		return nil
	case OpFinalize:
		return nil
	case OpSetupSession:
		return nil
	case OpTeardownSession:
		return nil
	case OpQuery:
		query := os.Getenv("ALBERT_QUERY")
		res, err := p.Query(query)
		if err != nil {
			return err
		}

		err = json.NewEncoder(p.Output).Encode(res)
		if err != nil {
			return err
		}
		return nil
	default:
		return AlbertError{
			Err:  fmt.Errorf("unknown albert op '%s'", op),
			Code: 255,
		}
	}
}
