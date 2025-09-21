package execx

import (
	"runtime"
	"testing"
)

func TestShell(t *testing.T) {
	sh, flag := Shell()
	if runtime.GOOS == "windows" {
		if sh != "cmd" || flag != "/C" {
			t.Fatalf("expected cmd /C on windows, got %s %s", sh, flag)
		}
	} else {
		if sh != "bash" || flag != "-lc" {
			t.Fatalf("expected bash -lc on unix, got %s %s", sh, flag)
		}
	}
}

func TestJoinCommand(t *testing.T) {
	if JoinCommand("echo", "hello", "world") != "echo hello world" {
		t.Fatalf("unexpected join result")
	}
}

func TestLookGo(t *testing.T) {
	if _, ok := Look("go"); !ok {
		t.Fatalf("expected to find 'go' in PATH during tests")
	}
}
