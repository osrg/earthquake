// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"

	mcli "github.com/mitchellh/cli"
	"github.com/osrg/namazu/nmz/cli/tools"
	coreutil "github.com/osrg/namazu/nmz/util/core"
)

type toolsCmd struct {
}

func (cmd toolsCmd) Help() string {
	return "Not documented yet."
}

func (cmd toolsCmd) Run(args []string) int {
	c := mcli.NewCLI("nmz tools", coreutil.NamazuVersion)
	c.Args = args
	c.Commands = map[string]mcli.CommandFactory{
		"visualize":  tools.VisualizeCommandFactory,
		"dump-trace": tools.DumpTraceCommandFactory,
		"summary":    tools.SummaryCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Printf("failed to execute search tool: %s\n", err)
	}

	return exitStatus
}

func (cmd toolsCmd) Synopsis() string {
	return "[Expert] Call miscellaneous tools related to `run` command"
}

func toolsCommandFactory() (mcli.Command, error) {
	return toolsCmd{}, nil
}
