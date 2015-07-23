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

package pbinspectorhandler

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
	"github.com/satori/go.uuid"
	. "../../equtils"
)

type PBInspectorHandler struct {
	EmulateREST bool
}

func recvPBMsgViaChan(entity *TransitionEntity, eventReqRecv chan *InspectorMsgReq) {
	for {
		req := &InspectorMsgReq{}

		rerr := RecvMsg(entity.Conn, req)
		if rerr != nil {
			if rerr == io.EOF {
				Log("received EOF from transition entity :%s", entity.Id)
				return
			} else {
				Log("failed to recieve request (transition entity: %s): %s", entity.Id, rerr)
				return // TODO: error handling
			}
		}

		Log("received message from transition entity :%s", entity.Id)
		eventReqRecv <- req
	}
}

func sendPBMsgViaChan(entity *TransitionEntity, eventRspSend chan *InspectorMsgRsp) {
	for {
		rsp := <-eventRspSend
		serr := SendMsg(entity.Conn, rsp)
		if serr != nil {
			Log("failed to send response (transition entity: %s): %s", entity.Id, serr)
			return // TODO: error handling
		}
		if *rsp.Res == InspectorMsgRsp_END {
			Log("send routine end (transition entity :%s)", entity.Id)
			return
		}
	}
}

func (handler *PBInspectorHandler) makeEmulatedRESTEventFromPBMsg (entity *TransitionEntity, req *InspectorMsgReq) *Event {
	evParam := 	 EAParam{
		// please refer to JSON schema file for this format
		"type": "event",
		"class": "",
		"deferred": true,
		"process": entity.Id,
		"uuid": uuid.NewV4().String(),
		"option": map[string]interface{} {
		},
	}
	if *req.Event.Type == InspectorMsgReq_Event_FUNC_CALL {
		evParam["class"] = "FunctionCallEvent"
		evParam["option"].(map[string]interface{})["func_name"] = *req.Event.FuncCall.Name
	} else if *req.Event.Type == InspectorMsgReq_Event_FUNC_RETURN {
		evParam["class"] = "FunctionReturnEvent"
		evParam["option"].(map[string]interface{})["func_name"] = *req.Event.FuncReturn.Name
	} else {
		Panic("invalid type of event: %d", *req.Event.Type)
	}
	if *req.HasJavaSpecificFields == 1 {
		evParam["option"].(map[string]interface{})["thread_name"] = *req.JavaSpecificFields.ThreadName
		stackTrace := make([]map[string]interface{}, 0)
		for _, stackTraceElement := range req.JavaSpecificFields.StackTraceElements {
			element := map[string]interface{}{
				"line_number": int(*stackTraceElement.LineNumber),
				"class_name":  *stackTraceElement.ClassName,
				"method_name": *stackTraceElement.MethodName,
				"file_name":   *stackTraceElement.FileName,
			}
			stackTrace = append(stackTrace, element)
		}
		evParam["option"].(map[string]interface{})["stack"] = stackTrace
		for _, param := range req.JavaSpecificFields.Params {
			evParam["option"].(map[string]interface{})[*param.Name] = *param.Value
		}
	}
	e := &Event{
		ArrivedTime: time.Now(),
		ProcId:      entity.Id,

		EventType:   "_JSON",
		EventParam: evParam,
	}
	return e
}

func (handler *PBInspectorHandler) makeEventFromPBMsg (entity *TransitionEntity, req *InspectorMsgReq) *Event {
	if ( handler.EmulateREST) {
		return handler.makeEmulatedRESTEventFromPBMsg(entity, req)
	}
	evType := ""
	evParam := NewEAParam()
	if *req.Event.Type == InspectorMsgReq_Event_FUNC_CALL {
		evType = "FuncCall"
		evParam["name"] = *req.Event.FuncCall.Name
	} else if *req.Event.Type == InspectorMsgReq_Event_FUNC_RETURN {
		evType = "FuncReturn"
		evParam["name"] = *req.Event.FuncReturn.Name
	} else {
		Panic("invalid type of event: %d", *req.Event.Type)
	}

	e := &Event{
		ArrivedTime: time.Now(),
		ProcId:      entity.Id,

		EventType:  evType,
		EventParam: evParam,
	}

	if *req.HasJavaSpecificFields == 1 {
		ejs := Event_JavaSpecific{
			ThreadName: *req.JavaSpecificFields.ThreadName,
		}

		for _, stackTraceElement := range req.JavaSpecificFields.StackTraceElements {
			element := Event_JavaSpecific_StackTraceElement{
				LineNumber: int(*stackTraceElement.LineNumber),
				ClassName:  *stackTraceElement.ClassName,
				MethodName: *stackTraceElement.MethodName,
				FileName:   *stackTraceElement.FileName,
			}

			ejs.StackTraceElements = append(ejs.StackTraceElements, element)
		}
		ejs.NrStackTraceElements = int(*req.JavaSpecificFields.NrStackTraceElements)

		for _, param := range req.JavaSpecificFields.Params {
			param := Event_JavaSpecific_Param{
				Name:  *param.Name,
				Value: *param.Value,
			}

			ejs.Params = append(ejs.Params, param)
		}

		ejs.NrParams = int(*req.JavaSpecificFields.NrParams)

		e.JavaSpecific = &ejs
	}
	return e
}

func (handler *PBInspectorHandler) makePBMsgFromAction (entity *TransitionEntity, req *InspectorMsgReq, action *Action) *InspectorMsgRsp {
	if (action.ActionType == "Accept") ||
	((action.ActionType == "_JSON") &&
	(action.ActionParam["class"] == "AcceptDeferredEventAction")) {
		result := InspectorMsgRsp_ACK
		req_msg_id := *req.MsgId
		rsp := &InspectorMsgRsp{
			Res:   &result,
			MsgId: &req_msg_id,
		}
		return rsp
	}
	Panic("unsupported action %s", action)
	return nil
}

func (handler *PBInspectorHandler) handleEntity(entity *TransitionEntity, readyEntityCh chan *TransitionEntity) {
	eventReqRecv := make(chan *InspectorMsgReq)
	eventRspSend := make(chan *InspectorMsgRsp)
	go recvPBMsgViaChan(entity, eventReqRecv)
	go sendPBMsgViaChan(entity, eventRspSend)

	for {
		select {
		case req := <-eventReqRecv:
			if *req.Type != InspectorMsgReq_EVENT {
				Log("invalid message from transition entity %s, type: %d", entity.Id, *req.Type)
				os.Exit(1)
			}

			if *req.Event.Type == InspectorMsgReq_Event_EXIT {
				Log("process %v is exiting", entity)
				continue
			}

			if entity.Id == "uninitialized" {
				// initialize id with a member of event
				entity.Id = *req.ProcessId
			}
			Log("event message received from transition entity %s", entity.Id)

			e := handler.makeEventFromPBMsg(entity, req)
			go func(ev *Event) {
				entity.EventToMain <- ev
			}(e)
			readyEntityCh <- entity

			if *req.Event.Type != InspectorMsgReq_Event_EXIT {
				act := <-entity.ActionFromMain
				Log("execute action (type=\"%s\")", act.ActionType)
				rsp := handler.makePBMsgFromAction(entity, req, act)
				eventRspSend <- rsp
				Log("accepted the event message from process %v", entity)
			}
		} // select
	} // for
} // func

func (handler *PBInspectorHandler) StartAccept(readyEntityCh chan *TransitionEntity) {
	sport := fmt.Sprintf(":%d", 10000) // FIXME (config.GetInt("inspectorHandler.pb.port"))
	ln, lerr := net.Listen("tcp", sport)
	if lerr != nil {
		Log("failed to listen on port %d: %s", 10000, lerr)
		os.Exit(1)
	}

	for {
		conn, aerr := ln.Accept()
		if aerr != nil {
			Log("failed to accept on %v: %s", ln, aerr)
			os.Exit(1)
		}

		Log("accepted new connection: %v", conn)

		entity := new(TransitionEntity)
		entity.Id = "uninitialized"
		entity.Conn = conn
		entity.ActionFromMain = make(chan *Action)
		entity.EventToMain = make(chan *Event)

		go handler.handleEntity(entity, readyEntityCh)
	}
}

func NewPBInspectorHanlder(config *Config) *PBInspectorHandler {
	emulateREST := config.GetBool("inspectorhandler.pb.emulateREST")
	if emulateREST {
		Log("Emulating REST Inspector Handler")
	}
	return &PBInspectorHandler{
		EmulateREST: emulateREST,
	}
}
