package errorwrapper

import (
	"errors"
	"io"
	"testing"
)

var (
	sentinel = errors.New("sentinel")
	sinkErr  error
	sinkStr  string
	sinkEW   ErrorWrapper
)

func TestNew_defaultJoiner(t *testing.T) {
	ew := New(0).(*errWrapper)
	if ew.errJoiner != defaultErrJoiner {
		t.Fatalf("want %q got %q", defaultErrJoiner, ew.errJoiner)
	}
}

func TestNew_customJoiner(t *testing.T) {
	ew := New('/').(*errWrapper)
	if ew.errJoiner != '/' {
		t.Fatalf("want '/' got %q", ew.errJoiner)
	}
}

func TestNew_prefix(t *testing.T) {
	ew := New('.', "svc").(*errWrapper)
	if ew.prefix != "svc" {
		t.Fatalf("want %q got %q", "svc", ew.prefix)
	}
}

func TestNew_noPrefix(t *testing.T) {
	ew := New('.').(*errWrapper)
	if ew.prefix != "" {
		t.Fatalf("want empty prefix got %q", ew.prefix)
	}
}

func TestNewError_nilReturnsNil(t *testing.T) {
	e := New('.', "svc")
	if e.NewError(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestNewErrorString_emptyReturnsNil(t *testing.T) {
	e := New('.', "svc")
	if e.NewErrorString("") != nil {
		t.Fatal("expected nil")
	}
}

func TestNewError_plainError(t *testing.T) {
	e := New('.', "svc")
	got := e.NewError(sentinel).Error()
	want := "svc: sentinel"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewError_withMsg(t *testing.T) {
	e := New('.', "svc")
	got := e.NewError(sentinel, "op failed").Error()
	want := "svc: [op failed] sentinel"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewErrorString_basic(t *testing.T) {
	e := New('.', "svc")
	got := e.NewErrorString("something went wrong").Error()
	want := "svc: something went wrong"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewErrorString_withMsg(t *testing.T) {
	e := New('.', "svc")
	got := e.NewErrorString("something went wrong", "context").Error()
	want := "svc: [context] something went wrong"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewError_chainedWrappers(t *testing.T) {
	e1 := New('.', "a")
	e2 := New('.', "b")
	e3 := New('.', "c")
	err := e3.NewError(e2.NewError(e1.NewError(sentinel)))
	want := "c.b.a: sentinel"
	if got := err.Error(); got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewError_chainedWrappersWithMsg(t *testing.T) {
	e1 := New('.', "a")
	e2 := New('.', "b")
	err := e2.NewError(e1.NewErrorString("boom", "inner"), "outer")
	want := "b.a: [outer] boom"
	if got := err.Error(); got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewError_customJoinerPreserved(t *testing.T) {
	e1 := New('/', "x")
	e2 := New('/', "y")
	err := e2.NewError(e1.NewError(sentinel))
	want := "y/x: sentinel"
	if got := err.Error(); got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestNewErrorString_joinerPreserved(t *testing.T) {
	e1 := New('/', "x")
	e2 := New('/', "y")
	err := e2.NewError(e1.NewErrorString("boom"))
	want := "y/x: boom"
	if got := err.Error(); got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestError_noPrefix(t *testing.T) {
	e := New('.').(*errWrapper)
	inner := &errWrapper{err: sentinel, errJoiner: '.'}
	got := e.NewError(inner).Error()
	want := "sentinel"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestError_noMsg(t *testing.T) {
	e := New('.', "svc")
	got := e.NewError(sentinel).Error()
	want := "svc: sentinel"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestUnwrap_errorsIs(t *testing.T) {
	e := New('.', "svc")
	wrapped := e.NewError(io.ErrNoProgress)
	if !errors.Is(wrapped, io.ErrNoProgress) {
		t.Fatal("errors.Is failed to locate sentinel through chain")
	}
}

func TestUnwrap_errorsIsChained(t *testing.T) {
	e1 := New('.', "a")
	e2 := New('.', "b")
	wrapped := e2.NewError(e1.NewError(io.ErrUnexpectedEOF))
	if !errors.Is(wrapped, io.ErrUnexpectedEOF) {
		t.Fatal("errors.Is failed across multi-level chain")
	}
}

func TestNewError_wrapNonErrWrapper(t *testing.T) {
	e := New('.', "svc")
	got := e.NewError(io.ErrNoProgress).Error()
	want := "svc: " + io.ErrNoProgress.Error()
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func BenchmarkNewError_shallow(b *testing.B) {
	e := New('.', "svc")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkErr = e.NewError(sentinel)
	}
}

func BenchmarkNewError_deep(b *testing.B) {
	e1 := New('.', "a")
	e2 := New('.', "b")
	e3 := New('.', "c")
	err1 := e1.NewError(sentinel)
	err2 := e2.NewError(err1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkErr = e3.NewError(err2)
	}
}

func BenchmarkNewErrorString(b *testing.B) {
	e := New('.', "svc")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkErr = e.NewErrorString("something went wrong")
	}
}

func BenchmarkError_shallow(b *testing.B) {
	e := New('.', "svc")
	err := e.NewError(sentinel, "op failed")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkStr = err.Error()
	}
}

func BenchmarkError_deep(b *testing.B) {
	e1 := New('.', "a")
	e2 := New('.', "b")
	e3 := New('.', "c")
	err := e3.NewError(e2.NewError(e1.NewError(sentinel, "inner"), "mid"), "outer")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkStr = err.Error()
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sinkEW = New('.', "svc")
	}
}
