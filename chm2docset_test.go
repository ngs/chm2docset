package main

import (
	"database/sql"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func cleanTmp() {
	os.RemoveAll("tmp")
}

type Test struct {
	actual   interface{}
	expected interface{}
}

func (test Test) Compare(t *testing.T) {
	if test.expected != test.actual {
		t.Errorf(`Expected "%v" but got "%v"`, test.expected, test.actual)
	}
}

func (test Test) DeepEqual(t *testing.T) {
	if !reflect.DeepEqual(test.expected, test.actual) {
		t.Errorf(`Expected "%v" but got "%v"`, test.expected, test.actual)
	}
}

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		if si, e := os.Stat(source); e != nil {
			err = os.Chmod(dest, si.Mode())
		}
	}

	return err
}

// Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return &CustomError{"Source is not a directory"}
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return &CustomError{"Destination already exists"}
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}

// A struct for returning custom error messages
type CustomError struct {
	What string
}

// Returns the error message defined in What as a string
func (e *CustomError) Error() string {
	return e.What
}

func TestNewOptionsSourceUnspecified(t *testing.T) {
	os.Args = []string{"chm2docset"}
	opts := NewOptions()
	if opts != nil {
		t.Errorf("Expected nil but got %v", opts)
	}
}

func TestNewOptionsDefaults(t *testing.T) {
	os.Args = []string{"chm2docset", "/foo/bar/baz.chm"}
	opts := NewOptions()
	for _, test := range []Test{
		Test{opts.Platform, "unknown"},
		Test{opts.Outdir, "./"},
		Test{opts.SourcePath, "/foo/bar/baz.chm"},
	} {
		test.Compare(t)
	}
}

func TestNewOptionsSpecifyOptions(t *testing.T) {
	os.Args = []string{"chm2docset", "-platform", "mac", "-out", "/qux", "/foo/bar/baz.chm"}
	opts := NewOptions()
	for _, test := range []Test{
		Test{opts.Platform, "mac"},
		Test{opts.Outdir, "/qux"},
		Test{opts.SourcePath, "/foo/bar/baz.chm"},
	} {
		test.Compare(t)
	}
}

func TestSourceFilename(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
	}
	Test{opts.SourceFilename(), "baz.chm"}.Compare(t)
}

func TestBasename(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
	}
	Test{opts.Basename(), "baz"}.Compare(t)
}

func TestDocsetPath(t *testing.T) {
	initFlags()
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "/qux",
	}
	Test{opts.DocsetPath(), "/qux/baz.docset"}.Compare(t)
	Test{opts.ContentPath(), "/qux/baz.docset/Contents/Resources/Documents"}.Compare(t)
	Test{opts.DatabasePath(), "/qux/baz.docset/Contents/Resources/docSet.dsidx"}.Compare(t)
	Test{opts.PlistPath(), "/qux/baz.docset/Contents/Info.plist"}.Compare(t)
	initFlags()
	opts = &Options{
		SourcePath: "/foo/bar/baz.chm",
	}
	Test{opts.DocsetPath(), "baz.docset"}.Compare(t)
	initFlags()
	opts = &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "/qux/foo.docset",
	}
	Test{opts.DocsetPath(), "/qux/foo.docset"}.Compare(t)
	Test{opts.ContentPath(), "/qux/foo.docset/Contents/Resources/Documents"}.Compare(t)
	Test{opts.DatabasePath(), "/qux/foo.docset/Contents/Resources/docSet.dsidx"}.Compare(t)
	Test{opts.PlistPath(), "/qux/foo.docset/Contents/Info.plist"}.Compare(t)
}

func TestBundleIdentifier(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/我輩は Lorem ipsum dolor sit amet,?-",
		Outdir:     "/qux/foo.docset",
	}
	Test{opts.BundleIdentifier(), "io.ngs.documentation.Loremipsumdolorsitamet-"}.Compare(t)
}

func TestPlistContent(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "/qux/foo.docset",
	}
	Test{opts.PlistContent(), `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>dashIndexFilePath</key>
    <string>Welcome.htm</string>
    <key>CFBundleIdentifier</key>
    <string>io.ngs.documentation.baz</string>
    <key>CFBundleName</key>
    <string>baz</string>
    <key>DocSetPlatformFamily</key>
    <string></string>
    <key>isDashDocset</key>
    <true/>
  </dict>
</plist>`}.Compare(t)
}

func TestWritePlist(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "tmp/foo.docset",
	}
	opts.CreateDirectory()
	err := opts.WritePlist()
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	b, err := ioutil.ReadFile("tmp/foo.docset/Contents/Info.plist")
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	Test{string(b), `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>dashIndexFilePath</key>
    <string>Welcome.htm</string>
    <key>CFBundleIdentifier</key>
    <string>io.ngs.documentation.baz</string>
    <key>CFBundleName</key>
    <string>baz</string>
    <key>DocSetPlatformFamily</key>
    <string></string>
    <key>isDashDocset</key>
    <true/>
  </dict>
</plist>`}.Compare(t)
	opts.Clean()
	if stat, err := os.Stat("tmp/foo.docset"); os.IsExist(err) {
		t.Errorf("Expect not exist but exists %v", stat)
	}
	cleanTmp()
}

func TestExtractSource(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "tmp/foo.docset",
	}
	opts.CreateDirectory()
	os.Setenv("PATH", "_fixtures/bin:"+os.Getenv("PATH"))
	err := opts.ExtractSource()
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	b, err := ioutil.ReadFile("tmp/fixtureinput.txt")
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	Test{string(b), "/foo/bar/baz.chm tmp/foo.docset/Contents/Resources/Documents\n"}.Compare(t)
	cleanTmp()
}

func TestCreateDatabase(t *testing.T) {
	opts := &Options{
		SourcePath: "/foo/bar/baz.chm",
		Outdir:     "tmp/Sample.docset",
	}
	opts.Clean()
	CopyDir("_fixtures/Sample.docset", "tmp/Sample.docset")
	opts.CreateDatabase()
	db, _ := sql.Open("sqlite3", opts.DatabasePath())
	rows, _ := db.Query("SELECT * FROM searchIndex")
	columns, _ := rows.Columns()
	Test{columns, []string{"id", "name", "type", "path"}}.DeepEqual(t)
	grid := [][]string{}
	for rows.Next() {
		var id string
		var name string
		var indexType string
		var path string
		err := rows.Scan(&id, &name, &indexType, &path)
		if err != nil {
			t.Errorf("Got errror %v", err)
		}
		grid = append(grid, []string{id, name, indexType, path})
	}
	Test{grid, [][]string{
		{"1", "test 4", "Guide", "/sub/test4.htm"},
		{"2", "test 1", "Guide", "/test1.htm"},
		{"3", "test 2 yo", "Guide", "/test2.htm"},
	}}.DeepEqual(t)
	cleanTmp()
}
