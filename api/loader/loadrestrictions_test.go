// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRestrictionNone(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	root := filesys.ConfirmedDir("irrelevant")
	path := "whatever"
	p, err := RestrictionNone(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}
}

func TestRestrictionDominatedShallowly(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	root, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		t.Fatal(err)
	}
	defer fSys.RemoveAll(string(root))

	// Legal; regular file is dominated.
	t.Run("legal dominated regular file", func(t *testing.T) {
		path := filepath.Join(string(root), "file")
		if _, err := fSys.Create(path); err != nil {
			t.Fatal(err)
		}
		p, err := RestrictionDominatedShallowly(fSys, root, path)
		if err != nil {
			t.Fatal(err)
		}
		pathDir, pathFile, err := fSys.CleanedAbs(path)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := pathDir.Join(pathFile), p; want != got {
			t.Fatalf("expected '%s', got '%s'", want, got)
		}
	})

	pathTarget := filepath.Join(root.Join("elsewhere"), "target")
	if err := fSys.MkdirAll(filepath.Dir(pathTarget)); err != nil {
		t.Fatal(err)
	}
	defer fSys.RemoveAll(filepath.Dir(pathTarget))
	if _, err := fSys.Create(pathTarget); err != nil {
		t.Fatal(err)
	}

	pathTrusted := root.Join("trusted")
	if err := fSys.MkdirAll(pathTrusted); err != nil {
		t.Fatal(err)
	}
	trusted, err := filesys.ConfirmDir(fSys, pathTrusted)
	if err != nil {
		t.Fatal(err)
	}

	// Legal; dominated symbolic link points to non-dominated target file.
	t.Run("legal dominated symbolic link", func(t *testing.T) {
		pathLink := trusted.Join("link")
		if err := os.Symlink(pathTarget, pathLink); err != nil {
			t.Fatal(err)
		}
		defer fSys.RemoveAll(filepath.Dir(pathLink))
		p, err := RestrictionDominatedShallowly(fSys, trusted, pathLink)
		if err != nil {
			t.Fatal(err)
		}
		pathTargetDir, pathTargetFile, err := fSys.CleanedAbs(pathTarget)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := pathTargetDir.Join(pathTargetFile), p; want != got {
			t.Fatalf("expected '%s', got '%s'", want, got)
		}
	})

	for _, test := range []struct {
		description string
		pathLink    string
	}{
		// Illegal; symbolic link exists but is out of bounds.
		{
			description: "illegal non-dominated symbolic link",
			pathLink:    filepath.Join(root.Join("somewhere-else"), "link"),
		},
		// Illegal; symbolic link exists but uses backsteps in the path starting below the root to place
		// it out of bounds.
		{
			description: "illegal non-dominated symbolic link with backsteps",
			pathLink:    filepath.Join(pathTrusted, "..", "somewhere-else", "link"),
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			if err := fSys.MkdirAll(filepath.Dir(test.pathLink)); err != nil {
				t.Fatal(err)
			}
			defer fSys.RemoveAll(filepath.Dir(test.pathLink))
			if err := os.Symlink(pathTarget, test.pathLink); err != nil {
				t.Fatal(err)
			}
			_, err := RestrictionDominatedShallowly(fSys, trusted, test.pathLink)
			if err == nil {
				t.Fatal("should have an error")
			}
			if !strings.Contains(
				err.Error(),
				fmt.Sprintf("file '%s' is not in or below '%s'", test.pathLink, trusted)) {
				t.Fatalf("unexpected err: %s", err)
			}
		})
	}
}

func TestRestrictionRootOnly(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	root := filesys.ConfirmedDir(
		filesys.Separator + filepath.Join("tmp", "foo"))
	path := filepath.Join(string(root), "whatever", "beans")

	fSys.Create(path)
	p, err := RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Legal.
	path = filepath.Join(
		string(root), "whatever", "..", "..", "foo", "whatever", "beans")
	p, err = RestrictionRootOnly(fSys, root, path)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(
		string(root), "whatever", "beans")
	if p != path {
		t.Fatalf("expected '%s', got '%s'", path, p)
	}

	// Illegal; file exists but is out of bounds.
	path = filepath.Join(filesys.Separator+"tmp", "illegal")
	fSys.Create(path)
	_, err = RestrictionRootOnly(fSys, root, path)
	if err == nil {
		t.Fatal("should have an error")
	}
	if !strings.Contains(
		err.Error(),
		"file '/tmp/illegal' is not in or below '/tmp/foo'") {
		t.Fatalf("unexpected err: %s", err)
	}
}
