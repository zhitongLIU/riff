/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	kailtesting "github.com/projectriff/riff/pkg/testing/kail"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestApplicationTailOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ApplicationTailOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.ApplicationTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "since duration",
			Options: &commands.ApplicationTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1m",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid duration",
			Options: &commands.ApplicationTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1",
			},
			ExpectFieldError: cli.ErrInvalidValue("1", cli.SinceFlagName),
		},
	}

	table.Run(t)
}

func TestApplicationTailCommand(t *testing.T) {
	defaultNamespace := "default"
	applicationName := "my-application"
	application := &buildv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      applicationName,
		},
	}

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "show logs",
			Args: []string{applicationName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("ApplicationLogs", mock.Anything, application, cli.TailSinceDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				application,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name: "show logs since",
			Args: []string{applicationName, cli.SinceFlagName, "1h"},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("ApplicationLogs", mock.Anything, application, time.Hour, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				application,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name:        "missing application",
			Args:        []string{applicationName},
			ShouldError: true,
		},
		{
			Name: "kail error",
			Args: []string{applicationName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("ApplicationLogs", mock.Anything, application, cli.TailSinceDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				application,
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewApplicationTailCommand)
}
