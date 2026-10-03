package main

import (
	"errors"
	"strings"
	"testing"
)

func TestNonInteractiveRequested(t *testing.T) {
	cases := []struct {
		name string
		args []string
		env  string
		want bool
	}{
		{name: "yes long", args: []string{"--yes"}, want: true},
		{name: "yes short", args: []string{"-y"}, want: true},
		{name: "env", env: "1", want: true},
		{name: "default", want: false},
		{name: "other env", env: "true", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nonInteractiveRequested(tc.args, tc.env)
			if got != tc.want {
				t.Fatalf("nonInteractiveRequested(%v, %q) = %v, want %v", tc.args, tc.env, got, tc.want)
			}
		})
	}
}

func TestRunConfiguredTasksStopsOnFailure(t *testing.T) {
	m := newModel()
	var ran []string
	m.tasks = []installTask{
		{name: "ok", description: "ok step", execute: func(*model) error {
			ran = append(ran, "ok")
			return nil
		}},
		{name: "bad", description: "bad step", execute: func(*model) error {
			ran = append(ran, "bad")
			return errors.New("boom")
		}},
		{name: "later", description: "later step", execute: func(*model) error {
			ran = append(ran, "later")
			return nil
		}},
	}

	err := runConfiguredTasks(&m)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want boom", err)
	}
	if strings.Join(ran, ",") != "ok,bad" {
		t.Fatalf("ran %v, want ok then bad only", ran)
	}
}
