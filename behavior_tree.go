// Copyright 2022 Guan Jianchang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ai

type BNodeState uint8

const (
	BNODE_STAT_NOT_EXECUTE BNodeState = iota
	BNODE_STAT_EXECUTING
	BNODE_STAT_SUCC
	BNODE_STAT_FAIL
)

type BNodeType uint8

const (
	BNODE_TYPE_ACTION BNodeType = iota
	BNODE_TYPE_SEQUENCE
	BNODE_TYPE_SELECT
	BNODE_TYPE_PARALLEL
)

const (
	BTREE_ROOT_NODE_ID = 1
)

//========================
//     BehaviorNode
//========================
type BehaviorNode interface {
	GetID() uint32
	GetActionID() uint32
	GetType() BNodeType
	UpdateStep()
	GetStep() uint32
	GetMaxStep() uint32
	GetState() BNodeState
	IsCompleted() bool
	Execute()

	AddChild(child BehaviorNode)
	RemoveChild(child BehaviorNode)
	RemoveChildByID(nodeId uint32)
	GetChildByID(nodeId uint32) (BehaviorNode, bool)
}

//========================
//     BaseBehaviorNode
//========================
type BaseBehaviorNode struct {
	nodeId   uint32
	nodeType BNodeType
	actionId uint32
	state    BNodeState
	step     uint32
	maxStep  uint32
}

func NewBaseBehaviorNode(nodeId uint32, actionId uint32, maxStep uint32) *BaseBehaviorNode {
	return &BaseBehaviorNode{
		nodeId:   nodeId,
		nodeType: BNODE_TYPE_ACTION,
		actionId: actionId,
		state:    BNODE_STAT_NOT_EXECUTE,
		step:     0,
		maxStep:  maxStep,
	}
}

func (n *BaseBehaviorNode) GetID() uint32 {
	return n.nodeId
}

func (n *BaseBehaviorNode) GetActionID() uint32 {
	return n.actionId
}

func (n *BaseBehaviorNode) GetType() BNodeType {
	return n.nodeType
}

func (n *BaseBehaviorNode) UpdateStep() {
	n.step++
}

func (n *BaseBehaviorNode) GetStep() uint32 {
	return n.step
}

func (n *BaseBehaviorNode) GetMaxStep() uint32 {
	return n.maxStep
}

func (n *BaseBehaviorNode) SetState(stat BNodeState) {
	n.state = stat
}

func (n *BaseBehaviorNode) GetState() BNodeState {
	return n.state
}

func (n *BaseBehaviorNode) IsCompleted() bool {
	if n.state == BNODE_STAT_SUCC {
		return true
	}

	if n.state == BNODE_STAT_FAIL {
		return true
	}

	return false
}

func (n *BaseBehaviorNode) Execute()                       {}
func (n *BaseBehaviorNode) AddChild(child BehaviorNode)    {}
func (n *BaseBehaviorNode) RemoveChild(child BehaviorNode) {}
func (n *BaseBehaviorNode) RemoveChildByID(nodeId uint32)  {}
func (n *BaseBehaviorNode) GetChildByID(nodeId uint32) (BehaviorNode, bool) {
	return nil, false
}

//========================
//     ControlNode
//========================
type ControlNode struct {
	*BaseBehaviorNode
	subNodes []BehaviorNode
}

func NewControlNode(nodeId uint32, nodeType BNodeType) *ControlNode {
	n := &ControlNode{
		BaseBehaviorNode: NewBaseBehaviorNode(nodeId, 0, 0),
		subNodes:         make([]BehaviorNode, 0),
	}

	n.nodeType = nodeType
	return n
}

func (n *ControlNode) AddChild(child BehaviorNode) {
	if child == nil {
		return
	}

	n.subNodes = append(n.subNodes, child)
}

func (n *ControlNode) RemoveChild(child BehaviorNode) {
	if child == nil {
		return
	}

	for i, exist := range n.subNodes {
		if exist == child {
			n.subNodes = append(n.subNodes[:i], n.subNodes[i+1:]...)
			break
		}
	}
}

func (n *ControlNode) RemoveChildByID(nodeId uint32) {
	for i, exist := range n.subNodes {
		if exist.GetID() == nodeId {
			n.subNodes = append(n.subNodes[:i], n.subNodes[i+1:]...)
			break
		}
	}
}

func (n *ControlNode) GetChildByID(nodeId uint32) (BehaviorNode, bool) {
	for _, exist := range n.subNodes {
		if exist.GetID() == nodeId {
			return exist, true
		}
	}

	return nil, false
}

//========================
//     SequenceNode
//========================
type SequenceNode struct {
	*ControlNode
}

func NewSequenceNode(nodeId uint32) *SequenceNode {
	return &SequenceNode{
		ControlNode: NewControlNode(nodeId, BNODE_TYPE_SEQUENCE),
	}
}

func (n *SequenceNode) Execute() {
	if n.IsCompleted() {
		return
	}

	n.state = BNODE_STAT_EXECUTING

	childLen := len(n.subNodes)
	for i, child := range n.subNodes {
		if child.IsCompleted() {
			continue
		}

		child.Execute()
		if !child.IsCompleted() {
			break
		}

		if child.GetState() == BNODE_STAT_FAIL {
			n.state = BNODE_STAT_FAIL
		} else if i == childLen-1 {
			n.state = BNODE_STAT_SUCC
		}

		break
	}
}

//========================
//     SelectNode
//========================
type SelectNode struct {
	*ControlNode
}

func NewSelectNode(nodeId uint32) *SelectNode {
	return &SelectNode{
		ControlNode: NewControlNode(nodeId, BNODE_TYPE_SELECT),
	}
}

func (n *SelectNode) Execute() {
	if n.IsCompleted() {
		return
	}

	n.state = BNODE_STAT_EXECUTING

	childLen := len(n.subNodes)
	for i, child := range n.subNodes {
		if child.IsCompleted() {
			continue
		}

		child.Execute()
		if !child.IsCompleted() {
			break
		}

		if child.GetState() == BNODE_STAT_SUCC {
			n.state = BNODE_STAT_SUCC
		} else if i == childLen-1 {
			n.state = BNODE_STAT_FAIL
		}

		break
	}
}

//========================
//     ParallelNode
//========================
type ParallelNode struct {
	*ControlNode
}

func NewParallelNode(nodeId uint32) *ParallelNode {
	return &ParallelNode{
		ControlNode: NewControlNode(nodeId, BNODE_TYPE_PARALLEL),
	}
}

func (n *ParallelNode) Execute() {
	if n.IsCompleted() {
		return
	}

	n.state = BNODE_STAT_EXECUTING

	bFinish := true
	for _, child := range n.subNodes {
		if child.IsCompleted() {
			continue
		}

		child.Execute()
		if !child.IsCompleted() {
			bFinish = false
			continue
		}

		if child.GetState() == BNODE_STAT_FAIL {
			n.state = BNODE_STAT_FAIL
			break
		}
	}

	if bFinish && (n.state == BNODE_STAT_EXECUTING) {
		n.state = BNODE_STAT_SUCC
	}
}

//========================
//      BehaviorTree
//========================
type BehaviorTree struct {
	treeId   uint32
	rootNode BehaviorNode
}

func NewBehaviorTree(treeId uint32) *BehaviorTree {
	return &BehaviorTree{
		treeId:   treeId,
		rootNode: NewSequenceNode(BTREE_ROOT_NODE_ID),
	}
}

func (t *BehaviorTree) GetID() uint32 {
	return t.treeId
}

func (t *BehaviorTree) GetRootNode() BehaviorNode {
	return t.rootNode
}

func (t *BehaviorTree) Execute() {
	t.rootNode.Execute()
}

func (t *BehaviorTree) GetState() BNodeState {
	return t.rootNode.GetState()
}

func (t *BehaviorTree) IsCompleted() bool {
	return t.rootNode.IsCompleted()
}
