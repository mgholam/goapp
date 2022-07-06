package zipfile

import (
	"archive/zip"
	"io"
	"os"
	"time"
)

var (
// UniqueFiles = false
)

// Add an existing file to a zip file
func AddFile(zipfilename string, filename string) error {

	fni, err := os.Stat(filename)
	if err != nil {
		return err
	}
	hdr := &zip.FileHeader{
		Name:     filename,
		Modified: fni.ModTime(),
		Method:   zip.Deflate,
		Comment:  fni.ModTime().String(),
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := os.Stat(zipfilename); err != nil {
		return zipCreateAdd(zipfilename, filename, hdr, f)
	}
	return zipAppend(zipfilename, filename, hdr, f)
}

// Add a stream to a zip file
func Add(zipfilename string, filename string, r io.Reader) error {

	hdr := &zip.FileHeader{
		Name:     filename,
		Modified: time.Now(),
		Method:   zip.Deflate,
		Comment:  time.Now().String(),
	}

	if _, err := os.Stat(zipfilename); err != nil {
		return zipCreateAdd(zipfilename, filename, hdr, r)
	}
	return zipAppend(zipfilename, filename, hdr, r)
}

func zipAppend(zipfilename, filename string, hdr *zip.FileHeader, r io.Reader) error {
	zipReader, err := zip.OpenReader(zipfilename)
	if err != nil {
		return err
	}
	targetFile, err := os.Create(zipfilename + ".tmp")
	if err != nil {
		return err
	}
	targetZipWriter := zip.NewWriter(targetFile)

	for _, zipItem := range zipReader.File {
		zipItemReader, _ := zipItem.Open()
		header := zipItem.FileHeader
		targetItem, _ := targetZipWriter.CreateHeader(&header)
		io.Copy(targetItem, zipItemReader)
	}

	z, _ := targetZipWriter.CreateHeader(hdr)

	io.Copy(z, r)
	zipReader.Close()
	targetZipWriter.Close()
	targetFile.Close()

	// rename output zipfile
	os.Remove(zipfilename)
	os.Rename(zipfilename+".tmp", zipfilename)
	return nil
}

func zipCreateAdd(zipfilename, filename string, hdr *zip.FileHeader, r io.Reader) error {
	archive, err := os.Create(zipfilename)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()
	z, err := zipWriter.CreateHeader(hdr)
	if err != nil {
		return err
	}
	if _, err := io.Copy(z, r); err != nil {
		return err
	}
	return nil
}
