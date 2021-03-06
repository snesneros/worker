// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package linux

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-vela/worker/version"

	"github.com/go-vela/types/constants"
	"github.com/go-vela/types/library"
	"github.com/go-vela/types/pipeline"

	"github.com/drone/envsubst"
	"github.com/sirupsen/logrus"
)

// CreateStep prepares the step for execution.
func (c *client) CreateStep(ctx context.Context, ctn *pipeline.Container) error {
	// update engine logger with extra metadata
	logger := c.logger.WithFields(logrus.Fields{
		"step": ctn.Name,
	})

	ctn.Environment["BUILD_HOST"] = c.Hostname
	ctn.Environment["VELA_HOST"] = c.Hostname
	ctn.Environment["VELA_VERSION"] = version.Version.String()
	// TODO: remove hardcoded reference
	ctn.Environment["VELA_RUNTIME"] = "docker"
	ctn.Environment["VELA_DISTRIBUTION"] = "linux"

	// TODO: remove hardcoded reference
	if ctn.Name == "init" {
		return nil
	}

	logger.Debug("setting up container")
	// setup the runtime container
	err := c.Runtime.SetupContainer(ctx, ctn)
	if err != nil {
		return err
	}

	logger.Debug("injecting secrets")
	// inject secrets for step
	err = injectSecrets(ctn, c.Secrets)
	if err != nil {
		return err
	}

	logger.Debug("marshaling configuration")
	// marshal container configuration
	body, err := json.Marshal(ctn)
	if err != nil {
		return fmt.Errorf("unable to marshal configuration: %v", err)
	}

	// create substitute function
	subFunc := func(name string) string {
		env := ctn.Environment[name]
		if strings.Contains(env, "\n") {
			env = fmt.Sprintf("%q", env)
		}

		return env
	}

	logger.Debug("substituting environment")
	// substitute the environment variables
	subStep, err := envsubst.Eval(string(body), subFunc)
	if err != nil {
		return fmt.Errorf("unable to substitute environment variables: %v", err)
	}

	logger.Debug("unmarshaling configuration")
	// unmarshal container configuration
	err = json.Unmarshal([]byte(subStep), ctn)
	if err != nil {
		return fmt.Errorf("unable to unmarshal configuration: %v", err)
	}

	return nil
}

// PlanStep defines a function that prepares the step for execution.
func (c *client) PlanStep(ctx context.Context, ctn *pipeline.Container) error {
	var err error

	b := c.build
	r := c.repo

	// update engine logger with extra metadata
	logger := c.logger.WithFields(logrus.Fields{
		"step": ctn.Name,
	})

	// update the engine step object
	s := new(library.Step)
	s.SetName(ctn.Name)
	s.SetNumber(ctn.Number)
	s.SetStatus(constants.StatusRunning)
	s.SetStarted(time.Now().UTC().Unix())
	s.SetHost(ctn.Environment["VELA_HOST"])
	s.SetRuntime(ctn.Environment["VELA_RUNTIME"])
	s.SetDistribution(ctn.Environment["VELA_DISTRIBUTION"])

	logger.Debug("uploading step state")
	// send API call to update the step
	s, _, err = c.Vela.Step.Update(r.GetOrg(), r.GetName(), b.GetNumber(), s)
	if err != nil {
		return err
	}

	s.SetStatus(constants.StatusSuccess)

	// add a step to a map
	c.steps.Store(ctn.ID, s)

	// get the step log here
	logger.Debug("retrieve step log")
	// send API call to capture the step log
	l, _, err := c.Vela.Log.GetStep(r.GetOrg(), r.GetName(), b.GetNumber(), s.GetNumber())
	if err != nil {
		return err
	}

	// add a step log to a map
	c.stepLogs.Store(ctn.ID, l)

	return nil
}

// ExecStep runs a step.
func (c *client) ExecStep(ctx context.Context, ctn *pipeline.Container) error {
	// TODO: remove hardcoded reference
	if ctn.Name == "init" {
		return nil
	}

	b := c.build
	r := c.repo

	result, ok := c.stepLogs.Load(ctn.ID)
	if !ok {
		return fmt.Errorf("unable to get step log from client")
	}

	l := result.(*library.Log)

	// update engine logger with extra metadata
	logger := c.logger.WithFields(logrus.Fields{
		"step": ctn.Name,
	})

	logger.Debug("running container")
	// run the runtime container
	err := c.Runtime.RunContainer(ctx, c.pipeline, ctn)
	if err != nil {
		return err
	}

	// create new buffer for uploading logs
	logs := new(bytes.Buffer)
	go func() error {
		logger.Debug("tailing container")
		// tail the runtime container
		rc, err := c.Runtime.TailContainer(ctx, ctn)
		if err != nil {
			return err
		}
		defer rc.Close()

		// create new scanner from the container output
		scanner := bufio.NewScanner(rc)

		// scan entire container output
		for scanner.Scan() {
			// write all the logs from the scanner
			logs.Write(append(scanner.Bytes(), []byte("\n")...))

			// if we have at least 1000 bytes in our buffer
			if logs.Len() > 1000 {
				logger.Trace(logs.String())

				// update the existing log with the new bytes
				l.SetData(append(l.GetData(), logs.Bytes()...))

				logger.Debug("appending logs")
				// send API call to update the logs for the step
				l, _, err = c.Vela.Log.UpdateStep(r.GetOrg(), r.GetName(), b.GetNumber(), ctn.Number, l)
				if err != nil {
					return err
				}

				// flush the buffer of logs
				logs.Reset()
			}
		}
		logger.Trace(logs.String())

		// update the existing log with the last bytes
		l.SetData(append(l.GetData(), logs.Bytes()...))

		logger.Debug("uploading logs")
		// send API call to update the logs for the step
		l, _, err = c.Vela.Log.UpdateStep(r.GetOrg(), r.GetName(), b.GetNumber(), ctn.Number, l)
		if err != nil {
			return err
		}

		return nil
	}()

	// do not wait for detached containers
	if ctn.Detach {
		return nil
	}

	logger.Debug("waiting for container")
	// wait for the runtime container
	err = c.Runtime.WaitContainer(ctx, ctn)
	if err != nil {
		return err
	}

	logger.Debug("inspecting container")
	// inspect the runtime container
	err = c.Runtime.InspectContainer(ctx, ctn)
	if err != nil {
		return err
	}

	return nil
}

// DestroyStep cleans up steps after execution.
func (c *client) DestroyStep(ctx context.Context, ctn *pipeline.Container) error {
	// TODO: remove hardcoded reference
	if ctn.Name == "init" {
		return nil
	}

	// update engine logger with extra metadata
	logger := c.logger.WithFields(logrus.Fields{
		"step": ctn.Name,
	})

	logger.Debug("removing container")
	// remove the runtime container
	err := c.Runtime.RemoveContainer(ctx, ctn)
	if err != nil {
		return err
	}

	return nil
}
