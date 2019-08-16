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
- name: build
  dir: /some-checkout-dir
  trigger:
    pubsub:
      topic: git-trigger
  tasks:
  - cmd: git pull
    timeout: 2s
  - cmd: make`,
			expected: &Config{
				Subscriptions: []ConfigSubscription{
					{
						Name: "build",
						Dir:  "/some-checkout-dir",
						Trigger: &ConfigTrigger{
							PubSub: &PubSubTrigger{
								Topic: "git-trigger",
							},
						},
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

func TestConfigCheck(t *testing.T) {
	cases := []struct {
		doc string
		err string
	}{
		{
			doc: "subscriptions:",
			err: "no subscriptions",
		},
		{
			doc: `
subscriptions:
- foo: bar`,
			err: `subscription "0": no trigger`,
		},
		{
			doc: `
subscriptions:
- name: foo`,
			err: `subscription "foo": no trigger`,
		},
		{
			doc: `
subscriptions:
- name: foo
  trigger: {}`,
			err: `subscription "foo": empty trigger`,
		},
		{
			doc: `
subscriptions:
- name: foo
  trigger:
    pubsub:
      topic: bar`,
			err: `subscription "foo": no tasks`,
		},
		{
			doc: `
subscriptions:
- name: foo
  trigger:
    pubsub:
      topic: bar
  tasks:
  - foo: bar`,
			err: `subscription "foo": task 0: no command`,
		},
		{
			doc: `
subscriptions:
- name: foo
  trigger:
    pubsub:
      topic: bar
  tasks:
  - cmd: ls`,
			err: `<nil>`,
		},
	}

	for i, c := range cases {
		config, err := readConfig(strings.NewReader(c.doc))
		if err != nil {
			t.Errorf("%d: error parsing config: %v", i, err)
			continue
		}

		actualErr := config.Check()
		actualMessage := "<nil>"
		if actualErr != nil {
			actualMessage = actualErr.Error()
		}

		if actualMessage != c.err {
			t.Errorf("%d\ngot:  %s\nwant: %s", i, actualMessage, c.err)
		}
	}
}
