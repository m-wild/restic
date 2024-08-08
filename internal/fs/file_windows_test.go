package fs_test

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/restic/restic/internal/fs"
	rtest "github.com/restic/restic/internal/test"
	"golang.org/x/sys/windows"
)

func TestTempFile(t *testing.T) {
	// create two temp files at the same time to check that the
	// collision avoidance works
	f, err := fs.TempFile("", "test")
	fn := f.Name()
	rtest.OK(t, err)
	f2, err := fs.TempFile("", "test")
	fn2 := f2.Name()
	rtest.OK(t, err)
	rtest.Assert(t, fn != fn2, "filenames don't differ %s", fn)

	_, err = os.Stat(fn)
	rtest.OK(t, err)
	_, err = os.Stat(fn2)
	rtest.OK(t, err)

	rtest.OK(t, f.Close())
	rtest.OK(t, f2.Close())

	_, err = os.Stat(fn)
	rtest.Assert(t, errors.Is(err, os.ErrNotExist), "err %s", err)
	_, err = os.Stat(fn2)
	rtest.Assert(t, errors.Is(err, os.ErrNotExist), "err %s", err)
}

func TestRecallOnDataAccessRealFile(t *testing.T) {
	// create a temp file for testing
	tempdir := rtest.TempDir(t)
	filename := filepath.Join(tempdir, "regular-file")
	err := os.WriteFile(filename, []byte("foobar"), 0640)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}

	xs := fs.ExtendedStat(fi)

	// ensure we can check attrs without error
	recall, err := fs.RecallOnDataAccess(xs)
	rtest.Assert(t, err == nil, "err should be nil", err)
	rtest.Assert(t, recall == false, "RecallOnDataAccess should be false")
}

// mockFileInfo implements os.FileInfo for mocking file attributes
type mockFileInfo struct {
	FileAttributes uint32
}

func (m mockFileInfo) IsDir() bool {
	return false
}
func (m mockFileInfo) ModTime() time.Time {
	return time.Now()
}
func (m mockFileInfo) Mode() iofs.FileMode {
	return 0
}
func (m mockFileInfo) Name() string {
	return "test"
}
func (m mockFileInfo) Size() int64 {
	return 0
}
func (m mockFileInfo) Sys() any {
	return &syscall.Win32FileAttributeData{
		FileAttributes: m.FileAttributes,
	}
}

func TestRecallOnDataAccessMockCloudFile(t *testing.T) {
	fi := mockFileInfo{
		FileAttributes: windows.FILE_ATTRIBUTE_RECALL_ON_DATA_ACCESS,
	}
	xs := fs.ExtendedStat(fi)

	recall, err := fs.RecallOnDataAccess(xs)
	rtest.Assert(t, err == nil, "err should be nil", err)
	rtest.Assert(t, recall, "RecallOnDataAccess should be true")
}

func TestRecallOnDataAccessMockRegularFile(t *testing.T) {
	fi := mockFileInfo{
		FileAttributes: windows.FILE_ATTRIBUTE_ARCHIVE,
	}
	xs := fs.ExtendedStat(fi)

	recall, err := fs.RecallOnDataAccess(xs)
	rtest.Assert(t, err == nil, "err should be nil", err)
	rtest.Assert(t, recall == false, "RecallOnDataAccess should be false")
}
