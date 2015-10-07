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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	. "./equtils"

	"./explorepolicy"
	"./historystorage"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"
)

var (
	workingDirPath   string
	materialsDirPath string

	config *Config
)

func init() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}

	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
}

func __createCmd(scriptPath, workingDir, materialsDir string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", scriptPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ() // this line is needed to extend current envs
	if workingDir != "" {  // can be empty for "init"
		cmd.Env = append(cmd.Env, "EQ_WORKING_DIR="+workingDir)
	}
	cmd.Env = append(cmd.Env, "EQ_MATERIALS_DIR="+materialsDir)

	return cmd
}

func createCmd(scriptPath string) *exec.Cmd {
	return __createCmd(scriptPath, workingDirPath, materialsDirPath)
}

func CreateKillCmd(entityId string) *exec.Cmd {
	if config.GetString("kill") == "" {
		return nil
	}

	killScriptPath := "\"" + materialsDirPath + "/" + config.GetString("kill") + "\"" + " " + entityId
	return createCmd(killScriptPath)
}

func run(args []string) {
	if len(args) != 1 {
		log.Critical("specify <storage dir path>")
		os.Exit(1) // panic() looks ugly for such an error
	}

	storagePath := args[0]
	confPath := storagePath + "/" + historystorage.StorageConfigPath

	err := error(nil)
	config, err = ParseConfigFile(confPath)
	if err != nil {
		log.Criticalf("failed to parse config file %s: %s", confPath, err)
		os.Exit(1)
	}

	storage := historystorage.New(config.GetString("storageType"), storagePath)
	storage.Init()

	policy := explorepolicy.CreatePolicy(config.GetString("explorePolicy"))
	if policy == nil {
		log.Criticalf("invalid policy name: %s", config.GetString("explorePolicy"))
		os.Exit(1)
	}
	policy.Init(storage, config.GetStringMap("explorePolicyParam"))

	workingDirPath = storage.CreateNewWorkingDir()
	InitLog(workingDirPath + "/earthquake.log")

	end := make(chan interface{})
	newTraceCh := make(chan *SingleTrace)

	go orchestrate(end, policy, newTraceCh, config)

	materialsDirPath = storagePath + "/" + storageMaterialsPath
	runScriptPath := materialsDirPath + "/" + config.GetString("run")

	cleanScriptPath := ""
	if config.GetString("clean") != "" {
		cleanScriptPath = materialsDirPath + "/" + config.GetString("clean")
	}

	validateScriptPath := ""
	if config.GetString("validate") != "" {
		validateScriptPath = materialsDirPath + "/" + config.GetString("validate")
	}

	runCmd := createCmd(runScriptPath)

	startTime := time.Now()

	log.Infof("Starting %s %s", runCmd.Path, runCmd.Args)
	rerr := runCmd.Run()
	log.Infof("Finished %s %s", runCmd.Path, runCmd.Args)
	if rerr != nil {
		panic(log.Criticalf("failed to execute run script %s: %s\n", runScriptPath, rerr))
	}

	log.Debug("Notifying finish to orchestrator")
	end <- true
	log.Debug("Waiting for trace from orchestrator")
	newTrace := <-newTraceCh
	log.Debug("Got the trace from orchestrator")

	endTime := time.Now()
	requiredTime := endTime.Sub(startTime)

	storage.RecordNewTrace(newTrace)

	succeed := true

	if validateScriptPath != "" {
		validateCmd := createCmd(validateScriptPath)

		rerr = validateCmd.Run()
		if rerr != nil {
			log.Infof("validation failed: %s", rerr)
			// TODO: detailed check of error
			// e.g. handle a case like permission denied, noent, etc
			storage.RecordResult(false, requiredTime)
			succeed = false
		} else {
			log.Info("validation succeed")
			storage.RecordResult(true, requiredTime)
		}
	}

	storage.Close()

	if !config.GetBool("notCleanIfValidationFail") || succeed {
		if cleanScriptPath != "" {
			cleanCmd := createCmd(cleanScriptPath)

			rerr = cleanCmd.Run()
			if rerr != nil {
				panic(log.Criticalf("failed to execute clean script %s: %s", cleanScriptPath, rerr))
			}
		}

		os.Exit(1)
	}
}

type runCmd struct {
}

func (cmd runCmd) Help() string {
	return "run help (todo)"
}

func (cmd runCmd) Run(args []string) int {
	run(args)
	return 0
}

func (cmd runCmd) Synopsis() string {
	return "run subcommand"
}

func runCommandFactory() (cli.Command, error) {
	return runCmd{}, nil
}
