package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	cases := []struct {
		doc      string
		expected *Config
	}{
		{
			doc: `
subscriptions:
- topic: git-trigger
  path: /some-checkout-dir
  tasks:
  - cmd: git pull
    timeout: 2s
  - cmd: make`,
			expected: &Config{
				Subscriptions: []ConfigSubscription{
					{
						Topic: "git-trigger",
						Path:  "/some-checkout-dir",
						Tasks: []ConfigTask{
							{Cmd: "git pull", Timeout: 2 * time.Second},
							{Cmd: "make"},
						},
					},
				},
			},
		},
	}

	for i, c := range cases {
		actual, err := readConfig(strings.NewReader(c.doc))
		if err != nil {
			t.Errorf("%d: error parsing config: %v", i, err)
			continue
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("%d\ngot:  %#v\nwant: %#v", i, actual, c.expected)
			continue
		}
	}
}
