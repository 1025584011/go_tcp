package libutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDirExists(t *testing.T) {
	root := os.TempDir()
	b, err := Exists(root)
	if err != nil {
		t.Error(err.Error())
	}
	if !b {
		t.Fail()
	}
}

func TestDirNotExists(t *testing.T) {
	root := os.TempDir()
	path := filepath.Join(root, "_TestDirNotExists")
	os.RemoveAll(path)
	b, err := Exists(path)
	if err != nil {
		t.Error(err.Error())
	}
	if b {
		t.Fail()
	}
}

func TestFileExists(t *testing.T) {
	root := os.TempDir()
	filename := filepath.Join(root, "_TestFileExists")
	if err := ioutil.WriteFile(filename, []byte{}, os.FileMode(0777)); err != nil {
		t.Error(err.Error())
	}
	b, err := Exists(filename)
	if err != nil {
		t.Error(err.Error())
	}
	if !b {
		t.Fail()
	}
}

func TestFileNotExists(t *testing.T) {
	root := os.TempDir()
	filename := filepath.Join(root, "_TestFileNotExists")
	os.Remove(filename)
	b, err := Exists(filename)
	if err != nil {
		t.Error(err.Error())
	}
	if b {
		t.Fail()
	}
}

func TestSymlinkExists(t *testing.T) {
	root := os.TempDir()
	filename := filepath.Join(root, "_TestSymlinkExists")
	if err := ioutil.WriteFile(filename, []byte{}, os.FileMode(0777)); err != nil {
		t.Error(err.Error())
	}
	link := filepath.Join(root, "_TestSymlinkExists.link")
	os.Remove(link)
	if err := os.Symlink(filename, link); err != nil {
		t.Error(err.Error())
	}
	b, err := Exists(link)
	if err != nil {
		t.Error(err.Error())
	}
	if !b {
		t.Fail()
	}
	os.Remove(filename)
	b, err = Exists(link)
	if err != nil {
		t.Error(err.Error())
	}
	if b {
		t.Fail()
	}
}

func TestIsLink(t *testing.T) {
	root := os.TempDir()
	filename := filepath.Join(root, "_TestIsLink")
	if err := ioutil.WriteFile(filename, []byte{}, os.FileMode(0777)); err != nil {
		t.Error(err.Error())
	}
	link := filepath.Join(root, "_TestIsLink.link")
	os.Remove(link)
	if err := os.Symlink(filename, link); err != nil {
		t.Error(err.Error())
	}
	b, err := IsLink(link)
	if err != nil {
		t.Error(err.Error())
	}
	if !b {
		t.Fail()
	}
	b, err = IsLink(filename)
	if err != nil {
		t.Error(err.Error())
	}
	if b {
		t.Fail()
	}
}

func TestIsEmpty(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "_TestIsEmpty")
	os.RemoveAll(dir)
	os.Mkdir(dir, os.FileMode(0777))
	b, err := IsEmpty(dir)
	if err != nil {
		t.Error(err.Error())
	}
	if !b {
		t.Fail()
	}
	b, err = IsEmpty(os.TempDir())
	if err != nil {
		t.Error(err.Error())
	}
	if b {
		t.Fail()
	}
}
