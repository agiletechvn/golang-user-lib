package dawg

import (
  "bufio"
  "fmt"
  "io"
  "sort"
)

// DAWG is a directed acyclic word graph.
type DAWG struct {
  root Node

  prevWord  string
  seq       int
  unchecked []unchecked
  minimized map[string]*Node
}

type unchecked struct {
  p *Node
  c *Node
  b byte
}

// FromReader constructs a DAWG from the newline-separated set of strings read from r.
// The strings do not need to be sorted.
func FromReader(r io.Reader) (*DAWG, error) {
  var words []string
  s := bufio.NewScanner(r)
  for s.Scan() {
    words = append(words, s.Text())
  }
  if err := s.Err(); err != nil {
    return nil, err
  }
  sort.Strings(words)
  var d DAWG
  for _, w := range words {
    d.Insert(w)
  }
  d.Finish()
  return &d, nil
}

// Root returns the root node of the graph.
func (d *DAWG) Root() *Node {
  return &d.root
}

func (d *DAWG) node() *Node {
  d.seq++
  return &Node{ID: d.seq}
}

func (d *DAWG) minimize(to int) {
  for i := len(d.unchecked) - 1; i >= to; i-- {
    n := d.unchecked[i]
    k := n.c.key()
    if m, ok := d.minimized[k]; ok {
      n.p.C[n.b] = m
    } else {
      if d.minimized == nil {
        d.minimized = make(map[string]*Node)
      }
      d.minimized[k] = n.c
    }
  }
  d.unchecked = d.unchecked[:to]
}

func min(a, b int) int {
  if a < b {
    return a
  }
  return b
}

func lcpLength(a, b string) int {
  n := min(len(a), len(b))
  for i := 0; i < n; i++ {
    if a[i] != b[i] {
      return i
    }
  }
  return n
}

// Insert inserts word into the DAWG.
// Insert must not be called after Finish is called.
func (d *DAWG) Insert(word string) {
  if word <= d.prevWord {
    panic("words must be inserted in alphabetical order")
  }

  lcpLen := lcpLength(word, d.prevWord)
  d.minimize(lcpLen)
  d.prevWord = word

  curr := &d.root
  if len(d.unchecked) > 0 {
    curr = d.unchecked[len(d.unchecked)-1].c
  }

  for i := lcpLen; i < len(word); i++ {
    b := word[i]
    next := d.node()
    if curr.C == nil {
      curr.C = make(map[byte]*Node)
    }
    curr.C[b] = next
    d.unchecked = append(d.unchecked, unchecked{curr, next, b})
    curr = next
  }

  curr.F = true
}

// Finish finishes the construction of the DAWG.
func (d *DAWG) Finish() {
  d.minimize(0)
}

// NodeCount returns the count of the nodes in the graph.
func (d *DAWG) NodeCount() int {
  return len(d.minimized) + 1
}

// NodeCount returns the count of the edges in the graph.
func (d *DAWG) EdgeCount() int {
  count := len(d.root.C)
  for _, n := range d.minimized {
    count += len(n.C)
  }
  return count
}

// Lookup searches for word in the DAWG. It returns true iff the DAWG contains the word.
func (d *DAWG) Lookup(word string) bool {
  curr := &d.root
  for j := 0; j < len(word); j++ {
    b := word[j]
    var ok bool
    curr, ok = curr.C[b]
    if !ok {
      return false
    }
  }

  return curr.F
}

// Lookup searches for word with the given prefix in the DAWG.
// Lookup returns the word if it is found, otherwise an empty string.
// The second return value indicates success.
func (d *DAWG) LookupPrefix(prefix string) (string, bool) {
  curr := &d.root
  for i := 0; i < len(prefix); i++ {
    b := prefix[i]
    c, ok := curr.C[b]
    if !ok {
      return "", false
    }
    curr = c
  }

  var buf []byte
  for !curr.F {
    for b, next := range curr.C {
      buf = append(buf, b)
      curr = next
      break
    }
  }

  return string(buf), true
}

func isPrint(b byte) bool {
  return b >= 0x20 && b <= 0x7e
}

// PrintDOT writes the DAWG to w in DOT format.
func (d *DAWG) PrintDOT(w io.Writer) error {
  fmt.Fprintln(w, "digraph g {")
  fmt.Fprintln(w, "  node [shape=circle];")
  printed := make(map[int]bool)
  nodes := []*Node{&d.root}
  for len(nodes) > 0 {
    n := nodes[0]
    nodes = nodes[1:]
    style := ""
    if n.F {
      style = ` style=filled fillcolor="#ff8080"`
    } else if n.ID == 0 {
      style = ` style=filled fillcolor="#80ff80"`
    }
    fmt.Fprintf(w, "  N%d [label=%[1]d%s];\n", n.ID, style)
    for b, c := range n.C {
      var label string
      if b == '"' {
        label = `\"`
      } else if isPrint(b) {
        label = fmt.Sprintf("%c", b)
      } else {
        label = fmt.Sprintf("%02X", b)
      }
      fmt.Fprintf(w, "  N%d -> N%d [label=\"%s\"];\n", n.ID, c.ID, label)
      if !printed[c.ID] {
        nodes = append(nodes, c)
        printed[c.ID] = true
      }
    }
  }
  fmt.Fprintln(w, "}")
  return nil
}

// Flatten returns the flattened representation of the DAWG.
func (d *DAWG) Flatten() []ArrayNode {
  var result []ArrayNode
  var to []int
  var hasChildren []bool
  seen := make(map[int]bool)
  indexes := make(map[int]int)
  index := 0
  nodes := []*Node{&d.root}

  for i := 0; i < len(nodes); i++ {
    node := nodes[i]
    indexes[node.ID] = index

    keys := sortedKeys(node.C)
    for j, b := range keys {
      c := node.C[b]

      if !seen[c.ID] {
        nodes = append(nodes, c)
        seen[c.ID] = true
      }

      to = append(to, c.ID)
      hasChildren = append(hasChildren, len(c.C) > 0)
      result = append(result, ArrayNode{
        B:   b,
        F:   c.F,
        EOL: j == len(keys)-1,
      })
      index++
    }
  }

  for i := range result {
    if !hasChildren[i] {
      continue
    }
    result[i].Index = indexes[to[i]]
  }

  return result
}
