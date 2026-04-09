package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestVMRuntimeStart(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "missing qemu binary", path: t.TempDir(), wantErr: "qemu-system-x86_64"},
		{name: "qemu present returns scaffold error", path: fakeQemuPath(t), wantErr: "not implemented yet"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("PATH", tc.path)
			r := vmRuntime{}

			// Act.
			_, err := r.Start(context.Background(), Spec{RunID: "run-1"})

			// Assert.
			if err == nil {
				t.Fatal("Start() error = nil, want non-nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Start() error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func fakeQemuPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	name := "qemu-system-x86_64"
	if runtime.GOOS == "windows" {
		name += ".bat"
	}
	path := filepath.Join(dir, name)
	content := []byte("#!/bin/sh\nexit 0\n")
	if runtime.GOOS == "windows" {
		content = []byte("@echo off\r\nexit /b 0\r\n")
	}
	if err := os.WriteFile(path, content, 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return dir
}
