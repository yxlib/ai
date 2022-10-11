// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ai

import "errors"

var (
	ErrNameLenZero      = errors.New("len of name is 0")
	ErrStatNil          = errors.New("state is nil")
	ErrActNil           = errors.New("action is nil")
	ErrTranNil          = errors.New("transition is nil")
	ErrTranNotExist     = errors.New("transition not exist")
	ErrEvtEmpty         = errors.New("event is empty")
	ErrNoFirstStat      = errors.New("no first state")
	ErrNoOldStat        = errors.New("no old state")
	ErrFromStatNotExist = errors.New("from state not exist")
	ErrToStatNotExist   = errors.New("to state not exist")
)

type FSMState interface {
	GetName() string
	OnEnter(fromState string)
	OnUpdate(dt int64)
	OnExit(toState string)
}

type FSMAction interface {
	GetName() string
	DoAction(evt string, param ...interface{}) bool
}

type FSMTransition struct {
	From   string
	Event  string
	To     string
	Action string
}

func NewFSMTransition(from string, evt string, to string, action string) *FSMTransition {
	return &FSMTransition{
		From:   from,
		Event:  evt,
		To:     to,
		Action: action,
	}
}

type FSM struct {
	id             uint32
	state          string
	oldStates      []string
	mapName2State  map[string]FSMState
	mapName2Action map[string]FSMAction
	transitions    []*FSMTransition
}

func NewFSM(id uint32) *FSM {
	return &FSM{
		id:             id,
		state:          "",
		oldStates:      make([]string, 0),
		mapName2State:  make(map[string]FSMState),
		mapName2Action: make(map[string]FSMAction),
		transitions:    make([]*FSMTransition, 0),
	}
}

func (f *FSM) GetID() uint32 {
	return f.id
}

func (f *FSM) GetCurState() string {
	return f.state
}

func (f *FSM) AddState(name string, stat FSMState) error {
	if len(name) == 0 {
		return ErrNameLenZero
	}

	if stat == nil {
		return ErrStatNil
	}

	f.mapName2State[name] = stat
	return nil
}

func (f *FSM) RemoveState(name string) {
	_, ok := f.mapName2State[name]
	if ok {
		delete(f.mapName2State, name)
	}
}

func (f *FSM) GetState(name string) (FSMState, bool) {
	stat, ok := f.mapName2State[name]
	return stat, ok
}

func (f *FSM) AddAction(name string, act FSMAction) error {
	if len(name) == 0 {
		return ErrNameLenZero
	}

	if act == nil {
		return ErrActNil
	}

	f.mapName2Action[name] = act
	return nil
}

func (f *FSM) RemoveAction(name string) {
	_, ok := f.mapName2Action[name]
	if ok {
		delete(f.mapName2Action, name)
	}
}

func (f *FSM) GetAction(name string) (FSMAction, bool) {
	act, ok := f.mapName2Action[name]
	return act, ok
}

func (f *FSM) AddTransition(from string, evt string, to string, action string) error {
	if len(from) == 0 {
		return ErrFromStatNotExist
	}

	if len(evt) == 0 {
		return ErrEvtEmpty
	}

	if len(to) == 0 {
		return ErrToStatNotExist
	}

	tran := NewFSMTransition(from, evt, to, action)
	f.transitions = append(f.transitions, tran)
	return nil
}

func (f *FSM) RemoveTransition(from string, evt string) {
	if len(from) == 0 {
		return
	}

	if len(evt) == 0 {
		return
	}

	for i, tran := range f.transitions {
		if tran.From == from && tran.Event == evt {
			f.transitions = append(f.transitions[:i], f.transitions[i+1:]...)
			break
		}
	}
}

func (f *FSM) GetTransition(from string, evt string) (*FSMTransition, bool) {
	if len(from) == 0 {
		return nil, false
	}

	if len(evt) == 0 {
		return nil, false
	}

	for _, tran := range f.transitions {
		if tran.From == from && tran.Event == evt {
			return tran, true
		}
	}

	return nil, false
}

func (f *FSM) Start(firstState string) error {
	if len(firstState) == 0 {
		return ErrNoFirstStat
	}

	stat, ok := f.GetState(firstState)
	if ok {
		f.state = firstState
		stat.OnEnter("")
	}
	return nil
}

func (f *FSM) Stop() {
	if len(f.state) == 0 {
		return
	}

	stat, ok := f.GetState(f.state)
	if ok {
		stat.OnExit("")
	}
}

func (f *FSM) Update(dt int64) {
	stat, ok := f.GetState(f.state)
	if ok {
		stat.OnUpdate(dt)
	}
}

func (f *FSM) Trigger(evt string, param ...interface{}) error {
	if len(evt) == 0 {
		return ErrEvtEmpty
	}

	if len(f.state) == 0 {
		return ErrNoFirstStat
	}

	triggerTran, ok := f.GetTransition(f.state, evt)
	if !ok {
		return ErrTranNotExist
	}

	// check transition
	oldStat, ok := f.GetState(f.state)
	if !ok {
		return ErrFromStatNotExist
	}

	newStat, ok := f.GetState(triggerTran.To)
	if !ok {
		return ErrToStatNotExist
	}

	// do transition
	act, ok := f.GetAction(triggerTran.Action)
	if ok {
		succ := act.DoAction(evt, param...)
		if !succ {
			return nil
		}
	}

	oldStat.OnExit(triggerTran.To)
	newStat.OnEnter(f.state)

	f.oldStates = append(f.oldStates, f.state)
	f.state = triggerTran.To
	return nil
}

func (f *FSM) PopState() error {
	if len(f.oldStates) == 0 {
		return ErrNoOldStat
	}

	oldStat, ok := f.GetState(f.state)
	if !ok {
		return ErrFromStatNotExist
	}

	idx := len(f.oldStates) - 1
	newStat, ok := f.GetState(f.oldStates[idx])
	if !ok {
		return ErrToStatNotExist
	}

	oldStat.OnExit(f.oldStates[idx])
	newStat.OnEnter(f.state)

	f.state = f.oldStates[idx]
	f.oldStates = f.oldStates[:idx]
	return nil
}
