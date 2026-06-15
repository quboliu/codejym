package api

// defaultPracticeFilename 是新用户默认训练组里那份「典型原文」的文件名。
const defaultPracticeFilename = "quicksort.go"

// defaultPracticeFile 是默认训练组的经典原文——快速排序的地道 Go 实现。
// 选它的原因：它是计算机科学里最有影响力的算法之一（Tony Hoare, 1960），
// 短小、递归、细节丰富，非常适合通过逐字临摹来内化。约 100 行。
const defaultPracticeFile = `// Package main 是快速排序（quicksort）的一个自包含、地道的实现 ——
// 计算机科学中最有影响力的算法之一，由 Tony Hoare 于 1960 年提出。
// 它非常适合通过临摹来内化：短小、递归，且充满了微小却关键的细节。
//
// 核心思想是「分治」（divide and conquer）：
//
//	1. 从切片中选一个基准元素（pivot）。
//	2. 分区：重排切片，使所有小于 pivot 的元素排在它前面、大于的排在后面。
//	   这一步之后，pivot 就落到了它最终排好序的位置上。
//	3. 对 pivot 两侧的子切片递归地重复以上步骤。
//
// 平均时间复杂度：O(n log n)
// 最坏情况：    O(n^2)  —— 很少见，且可通过更好的基准选择来规避。
// 空间复杂度：  O(log n) —— 原地排序，只占用递归栈。
package main

import "fmt"

// Sort 把整个切片按升序原地排序。
func Sort(a []int) {
	quicksort(a, 0, len(a)-1)
}

// quicksort 对闭区间 a[lo..hi] 排序。
func quicksort(a []int, lo, hi int) {
	// 零个或一个元素的区间天然有序：这是递归的基准情形，
	// 漏掉它正是最经典的 bug。
	if lo >= hi {
		return
	}

	p := partition(a, lo, hi)
	quicksort(a, lo, p-1) // 排序 pivot 左侧
	quicksort(a, p+1, hi) // 排序 pivot 右侧
}

// partition 用 Lomuto 方案围绕基准重排 a[lo..hi]，并返回 pivot 的最终下标。
// 为降低在「已基本有序」输入上退化成 O(n^2) 的概率，它先取首、中、尾三者的
// 中位数作为基准。
func partition(a []int, lo, hi int) int {
	mid := lo + (hi-lo)/2
	medianOfThree(a, lo, mid, hi)
	a[mid], a[hi] = a[hi], a[mid] // 把选中的基准暂放到末尾
	pivot := a[hi]

	// i 标记边界：a[lo..i-1] 里的元素都 <= pivot。
	i := lo
	for j := lo; j < hi; j++ {
		if a[j] <= pivot {
			a[i], a[j] = a[j], a[i]
			i++
		}
	}

	// 把基准交换到它最终的位置——正好在两个分区之间。
	a[i], a[hi] = a[hi], a[i]
	return i
}

// medianOfThree 用三次比较把 a[x] <= a[y] <= a[z] 排好，使中位数落在下标 y。
// 一个小小的辅助函数，却能带来很大的不同。
func medianOfThree(a []int, x, y, z int) {
	if a[x] > a[y] {
		a[x], a[y] = a[y], a[x]
	}
	if a[y] > a[z] {
		a[y], a[z] = a[z], a[y]
	}
	if a[x] > a[y] {
		a[x], a[y] = a[y], a[x]
	}
}

// isSorted 判断 a 是否非降序——用于校验结果，也方便写个快速的性质测试。
func isSorted(a []int) bool {
	for i := 1; i < len(a); i++ {
		if a[i-1] > a[i] {
			return false
		}
	}
	return true
}

func main() {
	data := []int{9, 3, 7, 1, 8, 2, 6, 5, 4, 0, 3, 7}
	fmt.Println("before:", data)

	Sort(data)

	fmt.Println("after: ", data)
	fmt.Println("sorted:", isSorted(data))
}
`
