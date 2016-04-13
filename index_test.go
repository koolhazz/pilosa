package pilosa_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/umbel/pilosa"
)

// Ensure index can sync with a remote index.
func TestIndexSyncer_SyncIndex(t *testing.T) {
	// Create a local index.
	idx0 := MustOpenIndex()
	defer idx0.Close()

	// Create a remote index wrapped by an HTTP
	idx1 := MustOpenIndex()
	defer idx1.Close()
	s := NewServer()
	defer s.Close()
	s.Handler.Index = idx1.Index

	// Set data on the local index.
	f := idx0.MustCreateFragmentIfNotExists("d", "f", 0)
	if _, err := f.SetBit(0, 10, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(2, 20, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(120, 10, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(200, 4, nil, 0); err != nil {
		t.Fatal(err)
	}

	f = idx0.MustCreateFragmentIfNotExists("d", "f0", 1)
	if _, err := f.SetBit(9, SliceWidth+5, nil, 0); err != nil {
		t.Fatal(err)
	}

	// Set data on the remote index.
	f = idx1.MustCreateFragmentIfNotExists("d", "f", 0)
	if _, err := f.SetBit(0, 4000, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(3, 10, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(120, 10, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(350, 0, nil, 0); err != nil {
		t.Fatal(err)
	}

	f = idx1.MustCreateFragmentIfNotExists("y", "z", 3)
	if _, err := f.SetBit(10, (3*SliceWidth)+4, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(10, (3*SliceWidth)+5, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(10, (3*SliceWidth)+7, nil, 0); err != nil {
		t.Fatal(err)
	}

	// Set up syncer.
	syncer := pilosa.IndexSyncer{
		Index:  idx0.Index,
		Client: MustNewClient(s.Host()).Client,
	}
	if err := syncer.SyncIndex(); err != nil {
		t.Fatal(err)
	}

	// Verify data is the same on both nodes.
	for i, idx := range []*Index{idx0, idx1} {
		f := idx.Fragment("d", "f", 0)
		if a := f.Bitmap(0).Bits(); !reflect.DeepEqual(a, []uint64{10, 4000}) {
			t.Fatalf("unexpected bits(%d/0): %+v", i, a)
		} else if a := f.Bitmap(2).Bits(); !reflect.DeepEqual(a, []uint64{20}) {
			t.Fatalf("unexpected bits(%d/2): %+v", i, a)
		} else if a := f.Bitmap(3).Bits(); !reflect.DeepEqual(a, []uint64{10}) {
			t.Fatalf("unexpected bits(%d/3): %+v", i, a)
		} else if a := f.Bitmap(120).Bits(); !reflect.DeepEqual(a, []uint64{10}) {
			t.Fatalf("unexpected bits(%d/120): %+v", i, a)
		} else if a := f.Bitmap(200).Bits(); !reflect.DeepEqual(a, []uint64{4}) {
			t.Fatalf("unexpected bits(%d/200): %+v", i, a)
		} else if a := f.Bitmap(350).Bits(); !reflect.DeepEqual(a, []uint64{0}) {
			t.Fatalf("unexpected bits(%d/350): %+v", i, a)
		}

		f = idx.Fragment("d", "f0", 1)
		if a := f.Bitmap(9).Bits(); !reflect.DeepEqual(a, []uint64{SliceWidth + 5}) {
			t.Fatalf("unexpected bits(%d/d/f0): %+v", i, a)
		}

		f = idx.Fragment("y", "z", 3)
		if a := f.Bitmap(10).Bits(); !reflect.DeepEqual(a, []uint64{(3 * SliceWidth) + 4, (3 * SliceWidth) + 5, (3 * SliceWidth) + 7}) {
			t.Fatalf("unexpected bits(%d/y/z): %+v", i, a)
		}
	}
}

// Index is a test wrapper for pilosa.Index.
type Index struct {
	*pilosa.Index
}

// NewIndex returns a new instance of Index with a temporary path.
func NewIndex() *Index {
	path, err := ioutil.TempDir("", "pilosa-")
	if err != nil {
		panic(err)
	}
	return &Index{Index: pilosa.NewIndex(path)}
}

// MustOpenIndex creates and opens an index at a temporary path. Panic on error.
func MustOpenIndex() *Index {
	i := NewIndex()
	if err := i.Open(); err != nil {
		panic(err)
	}
	return i
}

// Close closes the index and removes all underlying data.
func (i *Index) Close() error {
	defer os.RemoveAll(i.Path())
	return i.Index.Close()
}

// MustCreateFragmentIfNotExists returns a given fragment. Panic on error.
func (i *Index) MustCreateFragmentIfNotExists(db, frame string, slice uint64) *Fragment {
	f, err := i.Index.CreateFragmentIfNotExists(db, frame, slice)
	if err != nil {
		panic(err)
	}
	return &Fragment{Fragment: f}
}