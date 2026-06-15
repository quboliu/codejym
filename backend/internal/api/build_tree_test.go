package api

import "testing"

func findNode(ns []fileNode, name string) *fileNode {
	for i := range ns {
		if ns[i].Name == name {
			return &ns[i]
		}
	}
	return nil
}

// TestBuildFileTreeNesting 验证嵌套目录正确成树（旧实现因 map 乱序+按值拷贝会丢子节点），
// 且目录占位文件 .keep 不出现在树里、空目录仍可见。
func TestBuildFileTreeNesting(t *testing.T) {
	files := []string{
		"uploads/u/a/README.md",
		"uploads/u/a/src/.keep",
		"uploads/u/a/src/main.go",
		"uploads/u/a/src/util/helper.go",
		"uploads/u/a/empty/.keep",
	}
	tree := buildFileTree(files, "uploads/u/a/")

	if findNode(tree, "README.md") == nil {
		t.Fatal("README.md 缺失（根级文件）")
	}
	src := findNode(tree, "src")
	if src == nil || !src.IsDir {
		t.Fatal("src 目录缺失")
	}
	if findNode(src.Children, ".keep") != nil {
		t.Error(".keep 不应出现在树里")
	}
	if findNode(src.Children, "main.go") == nil {
		t.Error("src/main.go 缺失（嵌套丢节点 bug）")
	}
	util := findNode(src.Children, "util")
	if util == nil || !util.IsDir {
		t.Fatal("src/util 目录缺失")
	}
	if findNode(util.Children, "helper.go") == nil {
		t.Error("src/util/helper.go 缺失（深层嵌套 bug）")
	}
	// 只含 .keep 的空目录仍应作为目录出现
	empty := findNode(tree, "empty")
	if empty == nil || !empty.IsDir {
		t.Error("empty 空目录应可见")
	}
	if len(empty.Children) != 0 {
		t.Errorf("empty 目录应为空，实得 %d 个子节点", len(empty.Children))
	}
}
