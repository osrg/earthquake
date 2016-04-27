// Copyright (C) 2014 Nippon Telegraph and Telephone Corporation.
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

package test

import (
	"github.com/osrg/namazu/nmz/signal"
	"github.com/stretchr/testify/assert"
	"testing"
)

// used only for testing
func NewNopEvent(t *testing.T, entityID string, value int) signal.Event {
	m := map[string]interface{}{"value": value}
	event, err := signal.NewNopEvent(entityID, m)
	assert.NoError(t, err)
	return event
}

// used only for testing
func NewPacketEvent(t *testing.T, entityID string, value int) signal.Event {
	m := map[string]interface{}{"value": value}
	event, err := signal.NewPacketEvent(entityID, entityID, entityID, m)
	assert.NoError(t, err)
	return event
}
