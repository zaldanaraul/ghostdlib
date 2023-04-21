package zipper

import(
  "os"
  "fmt"
  "archive/zip"
  "path/filepath"
  "io"
  "io/fs"
  "log"
)

func newErr(err error, msg string, args ...any) (error) {
  if err != nil {
    formattedMessage := fmt.Sprintf(msg, args...)
    newErr := fmt.Errorf(formattedMessage + "\nError message: %v\n", err)
    return newErr
  }

  return nil
}

type Zipper struct {
  logger *log.Logger
}

type optFunc func(* Zipper) (*Zipper, error)

func WithLogger(l *log.Logger) optFunc {
  return func(z * Zipper) (*Zipper, error) {
    (*z).logger = l
    return z, nil
  }
}

func NewZipper(opts ...optFunc) (*Zipper, error) {
  z := &Zipper{}
  z.logger = log.New(io.Discard, "", 0)
  for _, opt := range opts {
    var err error
    z, err = opt(z)
    if err != nil {
      return nil, newErr(err, "Unable to apply option %v", opt)
    }
  }
  return z, nil
}


// Zip will create a zip file at the location dst
// from the contents at the path provided in src
// If src is a file, it will put the single file at the root level of the zip archive
// If src is a directory, it will place all the contents of the
// directory at the root level of the zip archivve
func (z* Zipper) Zip(src string, dst string) (error) {

  // Create zip file at destination
  zipFile, err := os.Create(dst)
  z.logger.Printf("Creating file at %v\n", dst)
  if err != nil {
    return newErr(err, "Unable to create file at %v", dst)
  }
  defer zipFile.Close()

  // Create zip writer
  zipWriter := zip.NewWriter(zipFile)
  defer zipWriter.Close()

  // Walk through any subpaths starting at src
  walker := func(path string, info fs.FileInfo, err error) (error) {
    z.logger.Printf("Walking %v\n", path)
    if err != nil {
      return err
    }

    if info.IsDir() {
      return nil
    }

    // open source file 
    srcFile, err := os.Open(path)
    if err != nil {
      return newErr(err, "Unable to open file %v", path)
    }

    // Get path relative to src
    relPath, err := filepath.Rel(src, path)

    // Handle case where src is a file
    if relPath == "." {
      relPath = filepath.Base(src)
    }

    // create file in zipped archive
    dstFile, err := zipWriter.Create(relPath)
    if err != nil {
      return newErr(err, "Unable to create file %v in zipped archive %v", relPath, dst)
    }

    _, err = io.Copy(dstFile, srcFile)
    if err != nil {
      return newErr(err, "Unable to copy file from %v to %v in %v", path, relPath, dst)
    }
    return nil
  }

  err = filepath.Walk(src, walker)
  if err != nil {
    return newErr(err, "Error while walking through path %v", src)
  }

  return nil
}
