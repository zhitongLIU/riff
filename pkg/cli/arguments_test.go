/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

func TestArgs(t *testing.T) {
	tests := []struct {
		name  string
		items []cli.Arg
		args  []string
		err   error
	}{{
		name: "no args",
	}, {
		name: "single arity",
		args: []string{"my-arg"},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}, {
		name: "missing args",
		args: []string{},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return fmt.Errorf("should not be called")
				},
			},
		},
		err: fmt.Errorf("missing required argument(s)"),
	}, {
		name: "extra args",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg-1", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		err: fmt.Errorf("unknown command %q for %q", "my-arg-2", "args-test"),
	}, {
		name: "multiple single arity",
		args: []string{"my-arg", "other-arg"},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("other-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}, {
		name: "capture arity",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{"my-arg-1", "my-arg-2"}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}, {
		name: "capture arity, no args",
		args: []string{},
		items: []cli.Arg{
			{
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}, {
		name: "capture arity, after single arity",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg-1", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
			{
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{"my-arg-2"}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}, {
		name: "optional args",
		args: []string{},
		items: []cli.Arg{
			{
				Arity:    1,
				Optional: true,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return fmt.Errorf("should not be called")
				},
			},
		},
	}, {
		name: "ignored args",
		args: []string{"my-arg"},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return cli.ErrIgnoreArg
				},
			},
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				Args: cli.Args(
					test.items...,
				),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
		})
	}
}

func TestNameArg(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   string
		expected string
		err      error
	}{{
		name: "too few args",
		err:  fmt.Errorf("missing required argument(s)"),
	}, {
		name:     "name arg",
		args:     []string{"my-name"},
		expected: "my-name",
	}, {
		name:     "too many args",
		args:     []string{"my-name", "extra-arg"},
		expected: "my-name",
		err:      fmt.Errorf("unknown command %q for %q", "extra-arg", "args-test"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				Args: cli.Args(
					cli.NameArg(&test.actual),
				),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestNamesArg(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   []string
		expected []string
		err      error
	}{{
		name:     "no name",
		args:     []string{},
		expected: []string{},
	}, {
		name:     "single name",
		args:     []string{"my-name"},
		expected: []string{"my-name"},
	}, {
		name:     "multiple names",
		args:     []string{"my-name", "my-other-name"},
		expected: []string{"my-name", "my-other-name"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				Args: cli.Args(
					cli.NamesArg(&test.actual),
				),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestBareDoubleDashArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   []string
		expected []string
		err      error
	}{{
		name: "no args",
		args: []string{},
	}, {
		name: "no bare double dash",
		args: []string{"my-arg", "my-other-arg"},
	}, {
		name:     "no args after bare double dash",
		args:     []string{"my-arg", "my-other-arg", "--"},
		expected: []string{},
	}, {
		name:     "bare double dash",
		args:     []string{"my-arg", "my-other-arg", "--", "my-name", "my-other-name"},
		expected: []string{"my-name", "my-other-name"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				Args: cli.Args(
					cli.BareDoubleDashArgs(&test.actual),
				),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}
