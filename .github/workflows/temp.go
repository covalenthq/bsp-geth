import (
	"strings"
)

type Trie struct {
	Root *Node
}

type Node struct {
	Val   string
	Left  *Node
	Right *Node
}

func Constructor() Trie {
	return Trie{}
}

func (this *Trie) Insert(word string) {
	this.Root = insert(this.Root, word)
}

func insert(root *Node, word string) {
	if root == nil {
		return &Node{
			Val: word,
		}
	}
	if word < root.Val {
		root.LeftNode = insert(root.LeftNode, word)
	} else {
		root.RightNode = insert(root.RightNode, word)
	}
}

func (this *Trie) Search(word string) bool {
	return search(this.Root, word)
}

func search(root *Node, word string) {
	if root == nil {
		return false
	}

	if root.Val == word {
		return true
	}

	return search(root.Left) || search(root.Right)
}

func (this *Trie) StartsWith(prefix string) bool {
	return startsWith(this.Root, prefix)
}

func startsWith(root *Node, prefix string) bool {
	if root == nil {
		return false
	}
	if strings.Index(root.Val, prefix) == 0 {
		return true
	}
	return startsWith(root.Left, prefix) || startsWith(root.Right, prefix)
}

/**
 * Your Trie object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Insert(word);
 * param_2 := obj.Search(word);
 * param_3 := obj.StartsWith(prefix);
 */