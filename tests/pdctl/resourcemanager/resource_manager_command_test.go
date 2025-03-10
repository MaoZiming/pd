// Copyright 2023 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resourcemanager_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tikv/pd/pkg/mcs/resourcemanager/server"
	"github.com/tikv/pd/pkg/utils/typeutil"
	"github.com/tikv/pd/tests"
	"github.com/tikv/pd/tests/pdctl"
	pdctlCmd "github.com/tikv/pd/tools/pd-ctl/pdctl"
)

func TestResourceManagerSuite(t *testing.T) {
	suite.Run(t, new(testResourceManagerSuite))
}

type testResourceManagerSuite struct {
	suite.Suite
	ctx     context.Context
	cancel  context.CancelFunc
	cluster *tests.TestCluster
	pdAddr  string
}

func (s *testResourceManagerSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	cluster, err := tests.NewTestCluster(s.ctx, 1)
	s.Nil(err)
	s.cluster = cluster
	s.cluster.RunInitialServers()
	cluster.WaitLeader()
	s.pdAddr = cluster.GetConfig().GetClientURL()
}

func (s *testResourceManagerSuite) TearDownSuite() {
	s.cancel()
	s.cluster.Destroy()
}

func (s *testResourceManagerSuite) TestConfigController() {
	expectCfg := server.Config{}
	expectCfg.Adjust(nil)
	// Show controller config
	checkShow := func() {
		args := []string{"-u", s.pdAddr, "resource-manager", "config", "controller", "show"}
		output, err := pdctl.ExecuteCommand(pdctlCmd.GetRootCmd(), args...)
		s.Nil(err)

		actualCfg := server.ControllerConfig{}
		err = json.Unmarshal(output, &actualCfg)
		s.Nil(err)
		s.Equal(expectCfg.Controller, actualCfg)
	}

	// Check default config
	checkShow()

	// Set controller config
	args := []string{"-u", s.pdAddr, "resource-manager", "config", "controller", "set", "ltb-max-wait-duration", "1h"}
	output, err := pdctl.ExecuteCommand(pdctlCmd.GetRootCmd(), args...)
	s.Nil(err)
	s.Contains(string(output), "Success!")
	expectCfg.Controller.LTBMaxWaitDuration = typeutil.Duration{Duration: 1 * time.Hour}
	checkShow()

	args = []string{"-u", s.pdAddr, "resource-manager", "config", "controller", "set", "enable-controller-trace-log", "true"}
	output, err = pdctl.ExecuteCommand(pdctlCmd.GetRootCmd(), args...)
	s.Nil(err)
	s.Contains(string(output), "Success!")
	expectCfg.Controller.EnableControllerTraceLog = true
	checkShow()

	args = []string{"-u", s.pdAddr, "resource-manager", "config", "controller", "set", "write-base-cost", "2"}
	output, err = pdctl.ExecuteCommand(pdctlCmd.GetRootCmd(), args...)
	s.Nil(err)
	s.Contains(string(output), "Success!")
	expectCfg.Controller.RequestUnit.WriteBaseCost = 2
	checkShow()
}
