package critbit

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestCritBit(t *testing.T) {
	sz := 1000
	kv := make(map[string]int)
	keys := make([]string, 0, sz)
	b := make([]byte, 64)

	tr := New()

	var s string
	for i := 0; i < sz; i++ {
		rand.Read(b)
		s = string(b)
		kv[s] = i
		_, ok := tr.Insert(b, i)
		if !ok {
			t.Fatalf("failed to insert to the tree. key: %v, i: %d", b, i)
		}
		b = b[:]
	}
	if tr.Len() != len(kv) {
		t.Fatalf("tree size does not match an expected one. got: %d, expected: %d", tr.Len(), len(kv))
	}
	for s, i := range kv {
		keys = append(keys, s)
		b := []byte(s)
		v, ok := tr.Get(b)
		if !ok {
			t.Fatalf("failed to get a value from the tree. key: %v, i: %d", b, i)
		}
		n, ok := v.(int)
		if !ok {
			t.Fatalf("returned value is not 'int' type value. type: %T", v)
		}
		if n != i {
			t.Errorf("returned value does not match an expected one. got: %d, expected: %d", n, i)
		}
	}

	sort.Strings(keys)

	if k, v, ok := tr.Minimum(); !ok {
		t.Fatalf("failed to get the minimum key value from the tree")
	} else if !bytes.Equal(k, []byte(keys[0])) {
		t.Errorf("returned minimum key does not match an expected one. got: %v, expected %v", []byte(k), []byte(keys[0]))
	} else {
		n, ok := v.(int)
		if !ok {
			t.Fatalf("returned minimum value is not 'int' type value. type: %T", v)
		}
		if n != kv[keys[0]] {
			t.Errorf("returned minimum value does not match an expected one. got: %d, expected: %d", n, kv[keys[0]])
		}
	}

	if k, v, ok := tr.Maximum(); !ok {
		t.Fatalf("failed to get the maximum key value from the tree")
	} else if !bytes.Equal(k, []byte(keys[len(keys)-1])) {
		t.Errorf("returned maximum key does not match an expected one. got: %v, expected %v", []byte(k), []byte(keys[len(keys)-1]))
	} else {
		n, ok := v.(int)
		if !ok {
			t.Fatalf("returned maximum value is not 'int' type value. type: %T", v)
		}
		if n != kv[keys[len(keys)-1]] {
			t.Errorf("returned maximum value does not match an expected one. got: %d, expected: %d", n, kv[keys[len(keys)-1]])
		}
	}

	result := make([]string, 0, sz)
	tr.Walk(func(k []byte, v interface{}) bool {
		result = append(result, string(k))
		return false
	})
	if len(result) != len(keys) {
		t.Fatalf("returned an unexpected walk result len. got: %d, expected: %d", len(result), len(keys))
	}
	for i, k := range keys {
		if result[i] != k {
			t.Fatalf("returned an unexpected walk result. got: %v, expected: %v", []byte(result[i]), []byte(k))
		}
	}

	for k, v := range kv {
		orig, ok := tr.Delete([]byte(k))
		if !ok {
			t.Fatalf("failed to delete the key from the tree. key: %v", k)
		}
		n, ok := orig.(int)
		if !ok {
			t.Fatalf("returned removed value is not 'int' type value. type: %T", orig)
		}
		if n != v {
			t.Errorf("returned removed value does not match an expected one. got: %d, expected: %d", n, v)
		}

		_, ok = tr.Get([]byte(k))
		if ok {
			t.Fatalf("the key was not removed actually from the tree. key: %v", k)
		}
	}
	if tr.Len() != 0 {
		t.Fatalf("tree size does not match an expected one. got: %d, expected: %d", tr.Len(), 0)
	}
}

func TestLongestPrefix(t *testing.T) {
	type testCase struct {
		prefix   string
		expected string
	}

	keys := []string{
		"",
		"foo",
		"foobar",
		"foobarbaz",
		"foobarbazzip",
		"foozip",
	}

	tr := New()
	for _, k := range keys {
		tr.Insert([]byte(k), nil)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), len(keys))
	}

	for _, tc := range []testCase{
		{prefix: "a", expected: ""},
		{prefix: "abc", expected: ""},
		{prefix: "fo", expected: ""},
		{prefix: "foo", expected: "foo"},
		{prefix: "foob", expected: "foo"},
		{prefix: "foobar", expected: "foobar"},
		{prefix: "foobarba", expected: "foobar"},
		{prefix: "foobarbaz", expected: "foobarbaz"},
		{prefix: "foobarbazzi", expected: "foobarbaz"},
		{prefix: "foobarbazzip", expected: "foobarbazzip"},
		{prefix: "foozi", expected: "foo"},
		{prefix: "foozip", expected: "foozip"},
		{prefix: "foozipzap", expected: "foozip"},
	} {
		t.Run(fmt.Sprintf("longest prefix %s", string(tc.prefix)), func(t *testing.T) {
			k, _, ok := tr.LongestPrefix([]byte(tc.prefix))
			if !ok {
				t.Fatalf("failed to get the longest prefix from the tree. prefix: %s", tc.prefix)
			}
			if !bytes.Equal(k, []byte(tc.expected)) {
				t.Errorf("returned longest prefix key does not match an expected one. got: %s, expected %s", string(k), tc.expected)
			}
		})
	}
}

func TestWalkPrefix(t *testing.T) {
	type testCase struct {
		path     string
		expected []string
	}

	keys := []string{
		"foobar",
		"foo/bar/baz",
		"foo/baz/bar",
		"foo/zip/zap",
		"zipzap",
	}

	tr := New()
	for _, k := range keys {
		tr.Insert([]byte(k), nil)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), len(keys))
	}

	for _, tc := range []testCase{
		{path: "f", expected: []string{"foo/bar/baz", "foo/baz/bar", "foo/zip/zap", "foobar"}},
		{path: "foo", expected: []string{"foo/bar/baz", "foo/baz/bar", "foo/zip/zap", "foobar"}},
		{path: "foob", expected: []string{"foobar"}},
		{path: "foo/", expected: []string{"foo/bar/baz", "foo/baz/bar", "foo/zip/zap"}},
		{path: "foo/b", expected: []string{"foo/bar/baz", "foo/baz/bar"}},
		{path: "foo/ba", expected: []string{"foo/bar/baz", "foo/baz/bar"}},
		{path: "foo/bar", expected: []string{"foo/bar/baz"}},
		{path: "foo/bar/baz", expected: []string{"foo/bar/baz"}},
		{path: "foo/bar/bazoo", expected: []string{}},
		{path: "z", expected: []string{"zipzap"}},
	} {
		t.Run(fmt.Sprintf("prefix %s", string(tc.path)), func(t *testing.T) {
			result := []string{}
			tr.WalkPrefix([]byte(tc.path), func(k []byte, v interface{}) bool {
				result = append(result, string(k))
				return false
			})
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("returned an unexpected keys. got: %#v, expected: %#v", result, tc.expected)
			}
		})
	}
}

func TestWalkPath(t *testing.T) {
	type testCase struct {
		path     string
		expected []string
	}

	keys := []string{
		"foo",
		"foo/bar",
		"foo/bar/baz",
		"foo/baz/bar",
		"foo/zip/zap",
		"zipzap",
	}

	tr := New()
	for _, k := range keys {
		tr.Insert([]byte(k), nil)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), len(keys))
	}

	for _, tc := range []testCase{
		{path: "f", expected: []string{}},
		{path: "foo", expected: []string{"foo"}},
		{path: "foo/", expected: []string{"foo"}},
		{path: "foo/ba", expected: []string{"foo"}},
		{path: "foo/bar", expected: []string{"foo", "foo/bar"}},
		{path: "foo/bar/baz", expected: []string{"foo", "foo/bar", "foo/bar/baz"}},
		{path: "foo/bar/bazoo", expected: []string{"foo", "foo/bar", "foo/bar/baz"}},
		{path: "z", expected: []string{}},
	} {
		t.Run(fmt.Sprintf("path %s", string(tc.path)), func(t *testing.T) {
			result := []string{}
			tr.WalkPath([]byte(tc.path), func(k []byte, v interface{}) bool {
				result = append(result, string(k))
				return false
			})
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("returned an unexpected keys. got: %#v, expected: %#v", result, tc.expected)
			}
		})
	}
}

func TestUpdateExistingKeyValue(t *testing.T) {
	key := []byte("foo")
	tr := New()
	tr.Insert(key, 1)
	if tr.Len() != 1 {
		t.Fatalf("tree size does not match an expected one. got: %d, expected: %d", tr.Len(), 1)
	}

	v, found := tr.Get(key)
	if !found {
		t.Fatalf("failed to get a value from the tree. key: %v", key)
	}
	n, ok := v.(int)
	if !ok {
		t.Fatalf("returned value is not 'int' type value. type: %T", v)
	}
	if n != 1 {
		t.Errorf("returned value does not match an expected one. got: %d, expected: %d", n, 1)
	}

	tr.Insert(key, 2)
	if tr.Len() != 1 {
		t.Fatalf("tree size does not match an expected one. got: %d, expected: %d", tr.Len(), 1)
	}

	v, found = tr.Get(key)
	if !found {
		t.Fatalf("failed to get a value from the tree. key: %v", key)
	}
	n, ok = v.(int)
	if !ok {
		t.Fatalf("returned value is not 'int' type value. type: %T", v)
	}
	if n != 2 {
		t.Errorf("returned value does not match an expected one. got: %d, expected: %d", n, 2)
	}
}

func TestOperationOnEmptyTree(t *testing.T) {
	tr := New()
	if _, ok := tr.Get([]byte("foo")); ok {
		t.Errorf("get something but should be empty")
	}
	if _, ok := tr.Delete([]byte("foo")); ok {
		t.Errorf("delete something but should fail")
	}
	if _, _, ok := tr.Minimum(); ok {
		t.Errorf("get something minimum but should be empty")
	}
	if _, _, ok := tr.Maximum(); ok {
		t.Errorf("get something maximum but should be empty")
	}
	if _, _, ok := tr.LongestPrefix([]byte("foo")); ok {
		t.Errorf("get something longest prefix but should be empty")
	}
	result := make([]string, 0)
	fn := func(k []byte, v interface{}) bool {
		result = append(result, string(k))
		return false
	}
	tr.Walk(fn)
	if len(result) > 0 {
		t.Errorf("get something from Walk but should be empty")
	}
	result = result[:0]
	tr.WalkPrefix([]byte("foo"), fn)
	if len(result) > 0 {
		t.Errorf("get something from WalkPrefix but should be empty")
	}
	result = result[:0]
	tr.WalkPath([]byte("foo"), fn)
	if len(result) > 0 {
		t.Errorf("get something from WalkPath but should be empty")
	}
	if ok := tr.Clear(); ok {
		t.Errorf("clear something but should fail")
	}
}

func TestEmptyKey(t *testing.T) {
	tr := New()
	tr.Insert([]byte(""), 1)
	if tr.Len() != 1 {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), 1)
	}
	v, found := tr.Get([]byte(""))
	if !found {
		t.Fatalf("failed to get a value of an empty key from the tree.")
	}
	n, ok := v.(int)
	if !ok {
		t.Fatalf("returned value is not 'int' type value. type: %T", v)
	}
	if n != 1 {
		t.Errorf("returned value does not match an expected one. got: %d, expected: %d", n, 1)
	}
}

func TestBugCase1(t *testing.T) {
	keys := []string{
		"WJUg",
		"LLj",
		"",
		"XfaMxKt",
		"l7Om",
		"ASWB",
		"wd0vboO",
		"qbUEE",
		"wnTR",
		"TPxlH",
	}

	tr := New()
	for _, k := range keys {
		tr.Insert([]byte(k), nil)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), len(keys))
	}

	sort.Strings(keys)

	result := make([]string, 0, len(keys))
	tr.Walk(func(k []byte, v interface{}) bool {
		result = append(result, string(k))
		return false
	})
	if len(result) != len(keys) {
		t.Fatalf("returned an unexpected walk result len. got: %d, expected: %d", len(result), len(keys))
	}
	for i, k := range keys {
		if result[i] != k {
			t.Fatalf("returned an unexpected walk result. got: %s, expected: %s", result[i], k)
		}
	}
}

type bytesSlice [][]byte

func (p bytesSlice) Len() int      { return len(p) }
func (p bytesSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p bytesSlice) Less(i, j int) bool {
	if bytes.Compare(p[i], p[j]) < 0 {
		return true
	}
	return false
}

func TestContainingZeroBytes(t *testing.T) {
	keys := [][]byte{
		{1, 0, 1},
		{1},
		{0, 1, 1},
		{},
		{0, 0, 1},
		{1, 1},
		{1, 1, 1},
		{0, 1},
		{0, 1, 0},
		{0, 0},
		{0, 0, 0},
		{0},
	}

	tr := New()
	for _, k := range keys {
		tr.Insert(k, nil)
	}
	if tr.Len() != len(keys) {
		t.Fatalf("tree is not an expected size. got: %d, expected: %d", tr.Len(), len(keys))
	}

	sort.Sort(bytesSlice(keys))

	result := make([][]byte, 0, len(keys))
	tr.Walk(func(k []byte, v interface{}) bool {
		result = append(result, k)
		return false
	})
	if len(result) != len(keys) {
		t.Fatalf("returned an unexpected walk result len. got: %d, expected: %d", len(result), len(keys))
	}
	for i, k := range keys {
		if !bytes.Equal(result[i], k) {
			t.Fatalf("returned an unexpected walk result. got: %v, expected: %v", result[i], k)
		}
	}

	pref0 := make([][]byte, 0, len(keys))
	for _, e := range keys {
		if len(e) > 0 && e[0] == 0 {
			pref0 = append(pref0, e)
		}
	}

	result = result[:0]
	tr.WalkPrefix([]byte{0}, func(k []byte, v interface{}) bool {
		result = append(result, k)
		return false
	})
	if len(result) != len(pref0) {
		t.Fatalf("returned an unexpected walk result len. got: %d, expected: %d", len(result), len(pref0))
	}
	for i, k := range pref0 {
		if !bytes.Equal(result[i], k) {
			t.Fatalf("returned an unexpected walk result. got: %v, expected: %v", result[i], k)
		}
	}
}

func loadTestFileData(b *testing.B, fpath string) [][]byte {
	f, err := os.Open(fpath)
	if err != nil {
		b.Fatalf("failed to open a test data file. %s", err)
	}
	defer f.Close()

	var data [][]byte
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadBytes('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				data = append(data, line[:len(line)-1])
			} else {
				data = append(data, line)
			}
		}
		if err != nil {
			break
		}
	}
	return data
}

func BenchmarkTreeInsertWords(b *testing.B) {
	words := loadTestFileData(b, "testdata/words.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr := New()
		for _, w := range words {
			tr.Insert(w, w)
		}
	}
}

func BenchmarkTreeGetWords(b *testing.B) {
	words := loadTestFileData(b, "testdata/words.txt")
	tr := New()
	for _, w := range words {
		tr.Insert(w, w)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range words {
			tr.Get(w)
		}
	}
}

func BenchmarkTreeInsertUUIDs(b *testing.B) {
	uuids := loadTestFileData(b, "testdata/uuid.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr := New()
		for _, id := range uuids {
			tr.Insert(id, id)
		}
	}
}

func BenchmarkTreeGetUUIDs(b *testing.B) {
	uuids := loadTestFileData(b, "testdata/uuid.txt")
	tr := New()
	for _, id := range uuids {
		tr.Insert(id, id)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, id := range uuids {
			tr.Get(id)
		}
	}
}
