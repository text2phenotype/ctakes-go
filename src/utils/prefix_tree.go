package utils

type PrefixTreeNode struct {
	Text     string
	Parent   *PrefixTreeNode
	Children map[rune]*PrefixTreeNode
}

type PrefixTree struct {
	Root *PrefixTreeNode
}

type StringPrefixTree struct {
	Root StringPrefixTreeNode
}

func (pTree *StringPrefixTree) Add(tokens []string, text string) {
	if len(tokens) == 0 || len(text) == 0 {
		return
	}

	node := pTree.Root
	for _, token := range tokens {
		childNode, isOk := node.Children[token]
		if isOk {
			node = childNode
			continue
		}

		newNode := StringPrefixTreeNode{
			Parent:   &node,
			Children: make(map[string]StringPrefixTreeNode),
		}

		if node.Children == nil {
			node.Children = make(map[string]StringPrefixTreeNode)
		}
		node.Children[token] = newNode
		node = newNode
	}

	node.Text = text
}

type StringPrefixTreeNode struct {
	Text     string
	Parent   *StringPrefixTreeNode
	Children map[string]StringPrefixTreeNode
}
