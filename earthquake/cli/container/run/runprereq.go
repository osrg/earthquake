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

package run

import (
	"fmt"

	"github.com/osrg/namazu/nmz/util/config"
	cap "github.com/syndtr/gocapability/capability"
)

func checkPrerequisite(cfg config.Config) error {
	dummyPID := 0
	capInst, err := cap.NewPid(dummyPID)
	if err != nil {
		return err
	}

	if cfg.GetBool("container.enableEthernetInspector") {
		if !capInst.Get(cap.EFFECTIVE, cap.CAP_NET_ADMIN) {
			return fmt.Errorf("CAP_NET_ADMIN is needed.")
		}
		if !capInst.Get(cap.EFFECTIVE, cap.CAP_SYS_ADMIN) {
			return fmt.Errorf("CAP_SYS_ADMIN is needed.")
		}
	}

	if cfg.GetBool("container.enableProcInspector") {
		if !capInst.Get(cap.EFFECTIVE, cap.CAP_SYS_NICE) {
			return fmt.Errorf("CAP_SYS_NICE is needed.")
		}
	}

	return nil
}
