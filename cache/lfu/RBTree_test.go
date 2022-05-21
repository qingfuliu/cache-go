package lfu

import (
	"fmt"
	"testing"
)

//1,2,#,#,3,#,#,
//func buildRBTree(s *string) *treeNode {
//	i := strings.Index(*s, ",")
//	nums := (*s)[:i]
//	*s = (*s)[i+1:]
//	if nums == "#" {
//		return nil
//	}
//	Root := &treeNode{
//		key: nums,
//	}
//
//	Root.leftChild = buildRBTree(s)
//
//	if Root.leftChild != nil {
//		Root.leftChild.parent = Root
//	}
//
//	Root.rightChild = buildRBTree(s)
//	if Root.rightChild != nil {
//		Root.rightChild.parent = Root
//	}
//	return Root
//}

//func TestRBTreeNode(t *testing.T) {
//	s := "5,2,1,#,1,#,#,3,3,#,#,4,#,#,345,34,#,34,#,#,678,#,#,"
//	Root := buildRBTree(&s)
//	PrintRBTree(Root)
//	fmt.Println()
//	left := Root.leftChild
//	left.swap(Root)
//	PrintRBTree(left)
//}

func TestRBTreeInsert(t *testing.T) {
	tree := NewRBTree()
	tree.Insert("345", nil)
	tree.Insert("12", nil)
	tree.Insert("455", nil)
	tree.Insert("452", nil)
	tree.Insert("458", nil)
	tree.Insert("451", nil)
	tree.Insert("10", nil)
	tree.Insert("11", nil)
	tree.Insert("9", nil)
	tree.Insert("8", nil)
	PrintRBTree(tree.Root)

}
func TestRBTreeDelete(t *testing.T) {
	tree := NewRBTree()
	tree.LruInsert("1", nil)
	tree.LruInsert("2", nil)
	tree.LruInsert("3", nil)
	tree.LruInsert("4", nil)
	tree.LruInsert("5", nil)
	tree.LruInsert("6", nil)
	////printRBTree(tree.root)
	//tree.Delete("345")
	//tree.Delete("455")
	//tree.Delete("452")
	//tree.Delete("12")
	//tree.Delete("451")
	//tree.Delete("9")
	//tree.Delete("11")
	//tree.Delete("8")
	//tree.Delete("10")
	//tree.Delete("458")
	tree.LruAdjustWithName("1")
	PrintRBTree(tree.Root)
}

func TestRBTreeInsertLru(t *testing.T) {
	tree := NewRBTree()
	tree.LruInsert("1", nil)
	PrintRBTree(tree.Root)
	tree.LruInsert("2", nil)
	PrintRBTree(tree.Root)
	tree.LruInsert("3", nil)
	PrintRBTree(tree.Root)
	tree.LruInsert("4", nil)
	PrintRBTree(tree.Root)
	tree.LruInsert("5", nil)
	PrintRBTree(tree.Root)
	tree.LruInsert("6", nil)
	PrintRBTree(tree.Root)
	fmt.Println("============================ delete ==================================")
	////printRBTree(tree.root)

	tree.LruDelete(tree.Leftmost)
	PrintRBTree(tree.Root)
	fmt.Println(tree.Leftmost)

	tree.LruDelete(tree.Leftmost)
	PrintRBTree(tree.Root)
	fmt.Println(tree.Leftmost)

	tree.LruDelete(tree.Leftmost)
	PrintRBTree(tree.Root)
	fmt.Println(tree.Leftmost)

	tree.LruDelete(tree.Leftmost)
	PrintRBTree(tree.Root)
	fmt.Println(tree.Leftmost)

	tree.LruDelete(tree.Leftmost)
	PrintRBTree(tree.Root)
	fmt.Println(tree.Leftmost)
}
