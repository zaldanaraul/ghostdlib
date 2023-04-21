package zipper_test

import (
  "archive/zip"
  "testing"
  "os"
  "bufio"
  "io"
  "io/fs"
  "log"
  "github.com/stretchr/testify/require"
  "github.com/stretchr/testify/assert"
  "github.com/zaldanaraul/ghostdlib/zipper"
)

func TestZipCanZipSingleFile(t *testing.T) {
  // Create temporary directory
  tmpDir := t.TempDir()

  // Create file to zip
  fileToZip := tmpDir + "/test_file.txt"
  f, err := os.Create(fileToZip)
  require.NoError(t, err)
  defer f.Close()
  w := bufio.NewWriter(f)
  fileContents := "Test text"
  _, err = w.WriteString(fileContents)
  require.NoError(t, err)
  err = w.Flush()
  require.NoError(t, err)

  // call zip function
  zippedArchive := t.TempDir() + "/test.zip"
  l := log.New(os.Stdout, "TestZipCanZipSingleFile: ", log.Flags())
  z, err := zipper.NewZipper(zipper.WithLogger(l))
  require.NoError(t, err)
  err = z.Zip(fileToZip, zippedArchive)
  require.NoError(t, err)

  // Check that zip archive contains test file
  zipArchiveReader, err := zip.OpenReader(zippedArchive)
  require.NoError(t, err)
  require.NotNil(t, zipArchiveReader)
  assert.Len(t, zipArchiveReader.File, 1)
  assert.Equal(t, "test_file.txt", zipArchiveReader.File[0].FileHeader.Name)

  // Check contents of file
  fileReadCloser, err := zipArchiveReader.File[0].Open()
  require.NoError(t, err)
  byteContents, err := io.ReadAll(fileReadCloser)
  require.Equal(t, fileContents, string(byteContents))
}

func TestZipCanZipFlatDirectory(t *testing.T) {
  // Create temporary directory
  tmpDir := t.TempDir()

  // Create directory to zip 
  dirToZip := tmpDir + "/test_dir"
  err := os.Mkdir(dirToZip, fs.FileMode(0755))
  require.NoError(t, err)

  // Create files in directory
  filesToZip := []string{"file1", "file2"}
  for _, fileName := range filesToZip {
    file, err := os.Create(dirToZip + "/" + fileName)
    require.NoError(t, err)
    io.WriteString(file, "This is " + fileName)
    err = file.Close()
    require.NoError(t, err)
  }

  // call zip function
  zippedArchive := t.TempDir() + "/test.zip"
  l := log.New(os.Stdout, "TestZipCanZipFlatDirectory: ", log.Flags())
  z, err := zipper.NewZipper(zipper.WithLogger(l))
  require.NoError(t, err)
  err = z.Zip(dirToZip, zippedArchive)
  require.NoError(t, err)

  // Check that zip archive contains test files
  zipArchiveReader, err := zip.OpenReader(zippedArchive)
  require.NoError(t, err)
  require.NotNil(t, zipArchiveReader)
  assert.Len(t, zipArchiveReader.File, len(filesToZip))
  for _, fileName := range filesToZip {
    foundInSlice := 0
    for zippedFileKey, zippedFile := range(zipArchiveReader.File) {
      if fileName == zippedFile.Name {
        foundInSlice = 1
        // check contents of file
        fileReadCloser, err := zipArchiveReader.File[zippedFileKey].Open()
        require.NoError(t, err)
        byteContents, err := io.ReadAll(fileReadCloser)
        require.Equal(t, "This is " + fileName, string(byteContents))
        break
      }
    }
    if foundInSlice != 1 {
      t.Errorf("%v was not in zipped archive", fileName)
    }
  }
}
