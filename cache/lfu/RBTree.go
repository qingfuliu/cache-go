package lfu

import (
	"cache-go/byteString"
	"fmt"
)

func PrintRBTree(root *treeNode) {
	stack := make([]*treeNode, 0)
	cur := root
	for len(stack) != 0 || cur != nil {
		for cur != nil {
			s := "Red"
			if !cur.isRed() {
				s = "Black"
			}
			fmt.Printf(" %v frequent:%d color:%s ", cur.key, cur.frequent, s)
			stack = append(stack, cur)
			cur = cur.leftChild
		}
		cur = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		cur = cur.rightChild
	}
	//s := "Red"
	//if !cur.left.isRed() {
	//	s = "Black"
	//}
	//fmt.Printf(" %v frequent:%d color:%s ", cur.leftChild.key, cur.leftChild.frequent, s)
	fmt.Println()
}

type Color bool

var Red Color = false

var Black Color = true

type treeNode struct {
	key        string
	color      Color
	val        *byteString.ByteString
	frequent   int64
	parent     *treeNode
	leftChild  *treeNode
	rightChild *treeNode
}

func (node *treeNode) Size() int64 {
	return int64(node.val.Len())
}

func (node *treeNode) getParent() *treeNode {
	return node.parent
}

func (node *treeNode) getAncestor() *treeNode {
	if node.parent == nil {
		return nil
	}
	return node.parent.parent
}

func (node *treeNode) getUncle() *treeNode {
	ancestor := node.getAncestor()
	if ancestor == nil {
		return nil
	}
	if node.parent == ancestor.leftChild {
		return ancestor.rightChild
	}
	return ancestor.leftChild
}

func (node *treeNode) getFarNephew() *treeNode {
	bother := node.getBother()
	if bother == nil {
		return nil
	}
	if node.parent.leftChild == node {
		return bother.rightChild
	}
	return bother.leftChild
}

func (node *treeNode) getNearNephew() *treeNode {
	bother := node.getBother()
	if bother == nil {
		return nil
	}
	if node.parent.leftChild == node {
		return bother.leftChild
	}
	return bother.rightChild
}

func (node *treeNode) getBother() *treeNode {
	if node.parent == nil {
		return nil
	}
	if node == node.parent.leftChild {
		return node.parent.rightChild
	}
	return node.parent.leftChild
}

func (node *treeNode) isRed() bool {
	return node.color == Red
}

func (node *treeNode) isLeftChild() bool {
	if node.parent == nil {
		return false
	}
	return node == node.parent.leftChild
}

func (node *treeNode) isRightChild() bool {
	if node.parent == nil {
		return false
	}
	return node == node.parent.rightChild
}

func (node *treeNode) leftRotate() {
	right := node.rightChild

	right.parent = node.parent
	if node.isLeftChild() {
		node.parent.leftChild = right
	} else if node.parent != nil {
		node.parent.rightChild = right
	}

	node.rightChild = right.leftChild
	if node.rightChild != nil {
		node.rightChild.parent = node
	}

	node.parent = right
	right.leftChild = node

}
func (node *treeNode) rightRotate() {
	left := node.leftChild

	left.parent = node.parent
	if node.isLeftChild() {
		node.parent.leftChild = left
	} else if node.parent != nil {
		node.parent.rightChild = left
	}

	node.leftChild = left.rightChild
	if node.leftChild != nil {
		node.leftChild.parent = node
	}

	node.parent = left
	left.rightChild = node
}

func (node *treeNode) remove() {
	if node.isLeftChild() {
		node.parent.leftChild = nil
	} else if node.parent != nil {
		node.parent.rightChild = nil
	}

	node.parent = nil
}

func (node *treeNode) swap(temp *treeNode) {
	if temp == nil || node == temp {
		return
	}
	parent := node
	if node.parent == temp {
		temp, parent = parent, temp
	}

	parent.color, temp.color = temp.color, parent.color
	if parent.isLeftChild() {
		parent.parent.leftChild = temp
	} else if parent.parent != nil {
		parent.parent.rightChild = temp
	}

	if parent != temp.parent {
		if temp.isLeftChild() {
			temp.parent.leftChild = parent
		} else if temp.parent != nil {
			temp.parent.rightChild = parent
		}
	}

	if parent.leftChild != nil && parent.leftChild != temp {
		parent.leftChild.parent = temp
	}

	if temp.leftChild != nil {
		temp.leftChild.parent = parent
	}

	if parent.rightChild != nil && parent.rightChild != temp {
		parent.rightChild.parent = temp
	}
	if temp.rightChild != nil {
		temp.rightChild.parent = parent
	}

	if temp.parent != parent {
		parent.parent, temp.parent = temp.parent, parent.parent
		parent.leftChild, temp.leftChild = temp.leftChild, parent.leftChild
		parent.rightChild, temp.rightChild = temp.rightChild, parent.rightChild
	} else {
		parent.parent, temp.parent = temp, parent.parent
		tempLeft, tempRight := temp.leftChild, temp.rightChild
		if parent.leftChild == temp {
			temp.leftChild, temp.rightChild = parent, parent.rightChild
		} else {
			temp.rightChild, temp.leftChild = parent, parent.leftChild
		}
		parent.leftChild, parent.rightChild = tempLeft, tempRight
	}
}

type RBTree struct {
	Root     *treeNode
	Leftmost *treeNode
}

func NewRBTree() *RBTree {
	return &RBTree{}
}

func (tree *RBTree) Search(key string, isInsert bool) (bs *byteString.ByteString) {
	node := tree.search(key, false)
	if node != nil {
		bs = node.val
	}
	return
}

func (tree *RBTree) search(key string, isInsert bool) (cur *treeNode) {
	cur = tree.Root
	for cur != nil {
		if key == cur.key {
			return
		}
		if key > cur.key {
			if cur.rightChild == nil && isInsert {
				return
			}
			cur = cur.rightChild
		} else {
			if cur.leftChild == nil && isInsert {
				return
			}
			cur = cur.leftChild
		}
	}
	return nil
}

func (tree *RBTree) Insert(key string, val *byteString.ByteString) {
	parent := tree.search(key, true)
	if parent != nil && parent.key == key {
		parent.val = val
		return
	}
	node := &treeNode{
		key:    key,
		val:    val,
		parent: parent,
	}
	if parent == nil {
		node.color = Black
		tree.Root = node
		return
	}

	if key > parent.key {
		parent.rightChild = node
	} else {
		parent.leftChild = node
	}

	tree.InsertAdjust(node)
}

func (tree *RBTree) Delete(key string) {
	node := tree.search(key, false)
	if node != nil {
		tree.deleteSpecial(node)
	}
}

func (tree *RBTree) InsertAdjust(node *treeNode) {
	parent := node.parent
begin:
	if node == tree.Root {
		node.color = Black
		return
	}
	//case 1 parent is a black node
	if parent == nil || !parent.isRed() {
		return
	}
	//case 2 parent is a red node
	//		case 2.1 uncle is a red node
	uncle := node.getUncle()
	if uncle != nil && uncle.isRed() {
		ancestor := node.getAncestor()
		ancestor.color, parent.color, uncle.color = Red, Black, Black
		parent = ancestor.parent
		node = ancestor
		goto begin
	}
	//		case2.2 uncle is a black node
	if node.isLeftChild() != parent.isLeftChild() {
		if node.isRightChild() {
			parent.leftRotate()
		} else {
			parent.rightRotate()
		}
		node = parent
	}
	ancestor := node.parent.parent
	node.parent.color, ancestor.color = ancestor.color, node.parent.color
	if node.isLeftChild() {
		ancestor.rightRotate()
	} else {
		ancestor.leftRotate()
	}

	if ancestor == tree.Root {
		tree.Root = ancestor.parent
	}
}
func (tree *RBTree) findSubstitute(node *treeNode) (sub *treeNode) {
	if node == nil {
		return nil
	}

	sub = node
	if sub.leftChild != nil {
		sub = sub.leftChild
		for sub.rightChild != nil {
			sub = sub.rightChild
		}
		return
	}
	if sub.rightChild != nil {
		sub = sub.rightChild
		for sub.leftChild != nil {
			sub = sub.leftChild
		}

		if node == tree.Leftmost {
			tree.Leftmost = sub
		}

		return
	}

	if node == tree.Leftmost {
		tree.Leftmost = node.parent
	}

	return
}

func (tree *RBTree) deleteAdjust(node *treeNode) {

	//case 1 :node's sub has a child
	if node.leftChild != nil {
		node.leftChild.parent, node.leftChild.color = node.parent, Black
		if node.isLeftChild() {
			node.parent.leftChild = node.leftChild
		} else if node.parent != nil {
			node.parent.rightChild = node.leftChild
		}
		node.parent, node.leftChild = nil, nil
		return
	}
	if node.rightChild != nil {
		node.rightChild.parent, node.rightChild.color = node.parent, Black
		if node.isLeftChild() {
			node.parent.leftChild = node.rightChild
		} else if node.parent != nil {
			node.parent.rightChild = node.rightChild
		}
		node.parent, node.rightChild = nil, nil
		return
	}
begin:
	//case 2 : node's sub had no child
	//	case 2.1:node's sub is a rad
	if node.isRed() || node == tree.Root {
		return
	}
	//	case 2.2:node's sub is a black
	//	case 2.2.1 bother is a red node
	bother := node.getBother()
	if bother != nil && bother.isRed() {
		bother.color, node.parent.color = node.parent.color, bother.color
		if node.isLeftChild() {
			node.parent.leftRotate()
		} else {
			node.parent.rightRotate()
		}
		if node.parent == tree.Root {
			tree.Root = node.parent.parent
		}
		bother = node.getBother()
	}
	//	case 2.2.2 sub has a red node
	var Nephew *treeNode
case2:
	Nephew = node.getFarNephew()
	if Nephew != nil && Nephew.isRed() {
		Nephew.color = Black
		node.parent.color, bother.color = bother.color, node.parent.color
		if node.isLeftChild() {
			node.parent.leftRotate()
		} else {
			node.parent.rightRotate()
		}
		if node.parent == tree.Root {
			tree.Root = node.parent.parent
		}
		return
	}
	Nephew = node.getNearNephew() //nearNephew
	if Nephew != nil && Nephew.isRed() {
		Nephew.color, bother.color = bother.color, Nephew.color
		if node.isLeftChild() {
			bother.rightRotate()
		} else {
			bother.leftRotate()
		}

		bother = Nephew
		goto case2
	}

	// all bother children is black
	if node.parent.isRed() {
		node.parent.color = Black
		if bother != nil {
			bother.color = Red
		}
		return
	}
	bother.color = Red
	node = node.parent
	goto begin
}

func (tree *RBTree) deleteSpecial(node *treeNode) {
	substitute := tree.findSubstitute(node)
	if substitute == tree.Root {
		tree.Root = nil
		return
	} else if node == tree.Root {
		tree.Root = substitute
	}
	node.swap(substitute)
	tree.deleteAdjust(node)
	node.remove()
}

//=====================================for frequent========================================//
func (tree *RBTree) searchLru(frequent int64) (cur *treeNode) {
	cur = tree.Root
	for cur != nil {
		if frequent > cur.frequent {
			if cur.rightChild == nil {
				return
			}
			cur = cur.rightChild
		} else {
			if cur.leftChild == nil {
				return
			}
			cur = cur.leftChild
		}
	}
	return nil
}

func (tree *RBTree) LruInsert(key string, val *byteString.ByteString) (node *treeNode) {
	node = &treeNode{
		key:    key,
		val:    val,
		parent: tree.Leftmost,
	}

	if tree.Root == nil {
		node.color = Black
		tree.Leftmost = node
		tree.Root = node
		return
	}
	tree.Leftmost.leftChild = node
	node.parent = tree.Leftmost
	tree.Leftmost = node
	tree.InsertAdjust(node)
	return
}

func (tree *RBTree) LruAdjustWithName(key string) {
	node := tree.search(key, false)
	if node != nil {
		tree.LruAdjust(node)
	}
}

func (tree *RBTree) LruAdjust(node *treeNode) {
	tree.LruDelete(node)
	node.frequent++
	parent := tree.searchLru(node.frequent)

	node.parent = parent
	node.color = Red
	if tree.Leftmost == nil || tree.Leftmost.frequent >= node.frequent {
		tree.Leftmost = node
	}

	if parent == nil {
		tree.Root = node
		return
	}

	if parent.frequent >= node.frequent {
		parent.leftChild = node
	} else {
		parent.rightChild = node
	}
	tree.InsertAdjust(node)
}

func (tree *RBTree) LruDeleteWithName(key string) {
	node := tree.search(key, false)
	if node != nil {
		tree.LruDelete(node)
	}
}

func (tree *RBTree) LruDelete(node *treeNode) {

	substitute := tree.findSubstitute(node)
	if substitute == tree.Root {
		tree.Leftmost = nil
		tree.Root = nil
		return
	} else if node == tree.Root {
		tree.Root = substitute
	}

	node.swap(substitute)
	tree.deleteAdjust(node)
	node.remove()
}

func (tree *RBTree) Eliminate() (node *treeNode) {
	node = tree.Leftmost
	tree.LruDelete(node)
	return
}
