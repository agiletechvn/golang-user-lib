package dawg

import (
  "fmt"
  "sort"
  "strconv"
)

// Node is a DAWG node.
type Node struct {
  ID int
  F  bool           // final?
  C  map[byte]*Node // children
}

// ArrayNode is a node in a flattened DAWG representation.
type ArrayNode struct {
  Index int
  B     byte // byte
  F     bool // final?
  EOL   bool // end of list?
}

func numDigits(n int) int {
  num := 1
  for n > 9 {
    num++
    n /= 10
  }
  return num
}

func sortedKeys(m map[byte]*Node) []byte {
  keys := make([]byte, 0, len(m))
  for b := range m {
    keys = append(keys, b)
  }
  sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
  return keys
}

func (n *Node) key() string {
  size := 2
  for _, c := range n.C {
    size += numDigits(c.ID) + 1
  }
  size += len(n.C)

  buf := make([]byte, 0, size)
  if n.F {
    buf = append(buf, '1')
  } else {
    buf = append(buf, '0')
  }
  buf = append(buf, '_')

  keys := sortedKeys(n.C)
  buf = append(buf, string(keys)...)

  for _, b := range keys {
    buf = append(buf, '_')
    buf = strconv.AppendInt(buf, int64(n.C[b].ID), 10)
  }
  if len(buf) != size {
    panic(fmt.Sprintf("%d %d %s", len(buf), size, string(buf)))
  }
  return string(buf)
}
