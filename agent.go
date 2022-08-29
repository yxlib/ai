// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ai

import "errors"

const (
	AGENT_STATE_IDLE = "agent_state_idle"
)

//========================
//     AgentFsmState
//========================
type AgentFsmStateListener interface {
	OnEnterFsmState(state string, fromState string)
	OnUpdateFsmState(state string, dt int64)
	OnExitFsmState(state string, toState string)
}

type AgentFsmState struct {
	name     string
	listener AgentFsmStateListener
}

func NewAgentFsmState(name string, listener AgentFsmStateListener) *AgentFsmState {
	return &AgentFsmState{
		name:     name,
		listener: listener,
	}
}

func (s *AgentFsmState) GetName() string {
	return s.name
}

func (s *AgentFsmState) OnEnter(fromState string) {
	if s.listener != nil {
		s.listener.OnEnterFsmState(s.name, fromState)
	}
}

func (s *AgentFsmState) OnUpdate(dt int64) {
	if s.listener != nil {
		s.listener.OnUpdateFsmState(s.name, dt)
	}
}

func (s *AgentFsmState) OnExit(toState string) {
	if s.listener != nil {
		s.listener.OnExitFsmState(s.name, toState)
	}
}

//========================
//     AgentFsmAction
//========================
type AgentFsmActionListener interface {
	OnFsmAction(action string, evt string, param ...interface{}) bool
}

type AgentFsmAction struct {
	name     string
	listener AgentFsmActionListener
}

func NewAgentFsmAction(name string, listener AgentFsmActionListener) *AgentFsmAction {
	return &AgentFsmAction{
		name:     name,
		listener: listener,
	}
}

func (a *AgentFsmAction) GetName() string {
	return a.name
}

func (a *AgentFsmAction) DoAction(evt string, param ...interface{}) bool {
	if a.listener != nil {
		return a.listener.OnFsmAction(a.name, evt, param...)
	}

	return false
}

//========================
//   AgentBehaviorNode
//========================
type AgentBNodeListener interface {
	OnBNodeAction(node BehaviorNode, param ...interface{}) BNodeState
}

type AgentBNode struct {
	*BaseBehaviorNode
	listener AgentBNodeListener
	params   []interface{}
}

func NewAgentBNode(nodeId uint32, actionId uint32, maxStep uint32, listener AgentBNodeListener, param ...interface{}) *AgentBNode {
	return &AgentBNode{
		BaseBehaviorNode: NewBaseBehaviorNode(nodeId, actionId, maxStep),
		listener:         listener,
		params:           param,
	}
}

func (a *AgentBNode) Execute() {
	if a.listener != nil {
		stat := a.listener.OnBNodeAction(a, a.params...)
		a.SetState(stat)
	}
}

//========================
//         Agent
//========================
type AgentFsmStateEnterFunc func(fromState string)
type AgentFsmStateUpdateFunc func(dt int64)
type AgentFsmStateExitFunc func(toState string)
type AgentFsmActionFunc func(evt string, param ...interface{}) bool
type AgentBNodeActionFunc func(node BehaviorNode, param ...interface{}) BNodeState

type Agent interface {
	GetID() uint32
	Update(dt int64)
}

type BaseAgent struct {
	agentId               uint32
	fsm                   *FSM
	mapState2BTree        map[string]*BehaviorTree
	mapState2EnterFunc    map[string]AgentFsmStateEnterFunc
	mapState2UpdateFunc   map[string]AgentFsmStateUpdateFunc
	mapState2ExitFunc     map[string]AgentFsmStateExitFunc
	mapName2FsmActionFunc map[string]AgentFsmActionFunc
	mapId2BNodeActionFunc map[uint32]AgentBNodeActionFunc
}

func NewBaseAgent(agentId uint32) *BaseAgent {
	return &BaseAgent{
		agentId:               agentId,
		fsm:                   NewFSM(agentId, AGENT_STATE_IDLE),
		mapState2BTree:        make(map[string]*BehaviorTree),
		mapState2EnterFunc:    make(map[string]AgentFsmStateEnterFunc),
		mapState2UpdateFunc:   make(map[string]AgentFsmStateUpdateFunc),
		mapState2ExitFunc:     make(map[string]AgentFsmStateExitFunc),
		mapName2FsmActionFunc: make(map[string]AgentFsmActionFunc),
		mapId2BNodeActionFunc: make(map[uint32]AgentBNodeActionFunc),
	}
}

func (a *BaseAgent) GetID() uint32 {
	return a.agentId
}

func (a *BaseAgent) Update(dt int64) {
	a.fsm.Update(dt)
}

func (a *BaseAgent) AddState(name string, behaviorTree *BehaviorTree, enterFunc AgentFsmStateEnterFunc, updateFunc AgentFsmStateUpdateFunc, exitFunc AgentFsmStateExitFunc) error {
	if len(name) == 0 {
		return errors.New("state is nil")
	}

	stat := NewAgentFsmState(name, a)
	a.fsm.AddState(name, stat)
	a.mapState2BTree[name] = behaviorTree
	a.mapState2EnterFunc[name] = enterFunc
	a.mapState2UpdateFunc[name] = updateFunc
	a.mapState2ExitFunc[name] = exitFunc
	return nil
}

func (a *BaseAgent) RemoveState(name string) error {
	if len(name) == 0 {
		return errors.New("state is nil")
	}

	a.fsm.RemoveState(name)
	_, ok := a.mapState2BTree[name]
	if ok {
		delete(a.mapState2BTree, name)
	}

	_, ok = a.mapState2EnterFunc[name]
	if ok {
		delete(a.mapState2EnterFunc, name)
	}

	_, ok = a.mapState2UpdateFunc[name]
	if ok {
		delete(a.mapState2UpdateFunc, name)
	}

	_, ok = a.mapState2ExitFunc[name]
	if ok {
		delete(a.mapState2ExitFunc, name)
	}

	return nil
}

func (a *BaseAgent) AddAction(name string, actionFunc AgentFsmActionFunc) error {
	if len(name) == 0 {
		return errors.New("action is nil")
	}

	act := NewAgentFsmAction(name, a)
	a.fsm.AddAction(name, act)
	a.mapName2FsmActionFunc[name] = actionFunc
	return nil
}

func (a *BaseAgent) RemoveAction(name string) error {
	if len(name) == 0 {
		return errors.New("action is nil")
	}

	a.fsm.RemoveAction(name)
	_, ok := a.mapName2FsmActionFunc[name]
	if ok {
		delete(a.mapName2FsmActionFunc, name)
	}

	return nil
}

func (a *BaseAgent) AddTransition(name string, tran *FSMTransition) error {
	if len(name) == 0 {
		return errors.New("action is nil")
	}

	return a.fsm.AddTransition(name, tran)
}

func (a *BaseAgent) RemoveTransition(name string) error {
	if len(name) == 0 {
		return errors.New("action is nil")
	}

	a.fsm.RemoveTransition(name)
	return nil
}

func (a *BaseAgent) AddBNodeActionHandleFunc(actionId uint32, handleFunc AgentBNodeActionFunc) error {
	_, ok := a.mapId2BNodeActionFunc[actionId]
	if ok {
		return errors.New("handle func exist")
	}

	a.mapId2BNodeActionFunc[actionId] = handleFunc
	return nil
}

func (a *BaseAgent) OnEnterFsmState(state string, fromState string) {
	f, ok := a.mapState2EnterFunc[state]
	if ok {
		f(fromState)
	}
}

func (a *BaseAgent) OnUpdateFsmState(state string, dt int64) {
	f, ok := a.mapState2UpdateFunc[state]
	if ok {
		f(dt)
	} else {
		btree, ok := a.mapState2BTree[state]
		if ok {
			btree.Execute()
		}
	}
}

func (a *BaseAgent) OnExitFsmState(state string, toState string) {
	f, ok := a.mapState2ExitFunc[state]
	if ok {
		f(toState)
	}
}

func (a *BaseAgent) OnFsmAction(action string, evt string, param ...interface{}) bool {
	f, ok := a.mapName2FsmActionFunc[action]
	if ok {
		return f(evt, param...)
	}

	return false
}

func (a *BaseAgent) OnBNodeAction(node BehaviorNode, param ...interface{}) BNodeState {
	actionId := node.GetActionID()
	f, ok := a.mapId2BNodeActionFunc[actionId]
	if ok {
		return f(node, param...)
	}

	return BNODE_STAT_NOT_EXECUTE
}
