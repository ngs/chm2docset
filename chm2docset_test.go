package main

import (
	"os"
	"reflect"
	"testing"
)

type Test struct {
	expected interface{}
	actual   interface{}
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
