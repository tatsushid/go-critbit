package critbit

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

type strs []string

func (ss *strs) String() string {
	return strings.Join(*ss, " ")
}

func (ss *strs) Set(value string) error {
	*ss = append(*ss, value)
	return nil
}

func TestMain(m *testing.M) {
	var printDotF bool
	var adds, dels strs
	var randomF int

	flag.BoolVar(&printDotF, "printdot", false, "print graphviz dot of critbit tree and exit")
	flag.Var(&adds, "add", "add `key` to critbit tree. this can be used multiple times")
	flag.Var(&dels, "del", "delete `key` from critbit tree. this can be used multiple times")
	flag.IntVar(&randomF, "random", 0, "insert keys chosen at random up to specified `times`")
	flag.Parse()

	if printDotF {
		t := New()
		for _, e := range adds {
			t.Insert([]byte(e), nil)
		}
		for _, e := range dels {
			t.Delete([]byte(e))
		}
		if randomF > 0 {
			chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, 8)
			rand.Seed(time.Now().UnixNano())
			for i := 0; i < randomF; i++ {
				l := rand.Intn(8)
				for j := 0; j < l; j++ {
					b[j] = chars[rand.Intn(len(chars))]
				}
				t.Insert(b[:l], nil)
				fmt.Fprintf(os.Stderr, "add %q\n", string(b[:l]))
			}
		}
		t.printDot(os.Stdout)
		os.Exit(0)
	}

	os.Exit(m.Run())
}

func (t *Tree) printDotR(w io.Writer, p node, prev, i int) int {
	id := i
	i++
	switch n := p.(type) {
	case *eNode:
		fmt.Fprintf(w, "    n%d [label=%q];\n", id, string(n.key))
		if id > 0 {
			fmt.Fprintf(w, "    n%d -> n%d;\n", prev, id)
		}
	case *iNode:
		var bit int
		onbit := ^n.other
		if onbit>>4 != 0 {
			onbit >>= 4
			bit += 4
		}
		if onbit>>2 != 0 {
			onbit >>= 2
			bit += 2
		}
		if onbit>>1 != 0 {
			bit += 1
		}
		fmt.Fprintf(w, "    n%d [label=\"pos:%d, bit:%d\"];\n", id, n.pos, bit)
		if id > 0 {
			fmt.Fprintf(w, "    n%d -> n%d;\n", prev, id)
		}
		for j := 0; j < 2; j++ {
			if n.children[j] != nil {
				i = t.printDotR(w, n.children[j], id, i)
			}
		}
	}
	return i
}

func (t *Tree) printDot(w io.Writer) {
	fmt.Fprintf(w, "digraph critbit {\n")
	fmt.Fprintf(w, "    node [style=filled];\n")

	t.printDotR(w, t.root, 0, 0)

	fmt.Fprintf(w, "}\n")
}
