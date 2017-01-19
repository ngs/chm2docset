package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func failOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Options options
type Options struct {
	Outdir     string
	Platform   string
	SourcePath string
}

var platform string
var outdir string

func initFlags() {
	platform = "unknown"
	outdir = "./"
}

func init() {
	initFlags()
	flag.Usage = usage
	flag.StringVar(&platform, "platform", platform, "DocSet Platform Family")
	flag.StringVar(&outdir, "out", outdir, "Output directory or file path")
}

// NewOptions returns new options
func NewOptions() *Options {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		return nil
	}
	return &Options{
		Outdir:     outdir,
		Platform:   platform,
		SourcePath: args[0],
	}
}

// SourceFilename returns source file name
func (opts *Options) SourceFilename() string {
	return filepath.Base(opts.SourcePath)
}

// Basename returns file basename
func (opts *Options) Basename() string {
	fn := opts.SourceFilename()
	ext := filepath.Ext(fn)
	return fn[0 : len(fn)-len(ext)]
}

// DocsetPath returns path to docset bundle
func (opts *Options) DocsetPath() string {
	if strings.HasSuffix(opts.Outdir, ".docset") {
		return opts.Outdir
	}
	return path.Join(opts.Outdir, opts.Basename()+".docset")
}

// ContentPath returns path to docset resources
func (opts *Options) ContentPath() string {
	return path.Join(opts.DocsetPath(), "Contents", "Resources", "Documents")
}

// DatabasePath returns path to SQLite3 database
func (opts *Options) DatabasePath() string {
	return path.Join(opts.DocsetPath(), "Contents", "Resources", "docSet.dsidx")
}

// PlistPath returns path to Info.plist
func (opts *Options) PlistPath() string {
	return path.Join(opts.DocsetPath(), "Contents", "Info.plist")
}

// BundleIdentifier returns bundle identifier of docset bundle
func (opts *Options) BundleIdentifier() string {
	safeRE := regexp.MustCompile("[^^a-zA-Z\\d-_]")
	return "io.ngs.documentation." + safeRE.ReplaceAllString(opts.Basename(), "")
}

// PlistContent returns plsit content
func (opts *Options) PlistContent() string {
	// https://kapeli.com/resources/Info.plist
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>dashIndexFilePath</key>
    <string>Welcome.htm</string>
    <key>CFBundleIdentifier</key>
    <string>` + opts.BundleIdentifier() + `</string>
    <key>CFBundleName</key>
    <string>` + opts.Basename() + `</string>
    <key>DocSetPlatformFamily</key>
    <string>` + opts.Platform + `</string>
    <key>isDashDocset</key>
    <true/>
  </dict>
</plist>`
}

// WritePlist writes plist file
func (opts *Options) WritePlist() error {
	return ioutil.WriteFile(opts.PlistPath(), []byte(opts.PlistContent()), 644)
}

// Clean removes existing output
func (opts *Options) Clean() error {
	return os.RemoveAll(opts.DocsetPath())
}

// ExtractSource extracts source to destination
func (opts *Options) ExtractSource() error {
	if err := os.MkdirAll(opts.ContentPath(), 0755); err != nil {
		return err
	}
	cmd := exec.Command("extract_chmLib", opts.SourcePath, opts.ContentPath())
	return cmd.Run()
}

// CreateDatabase creates database
func (opts *Options) CreateDatabase() error {
	titleRE := regexp.MustCompile("<title>([^<]+)</title>")
	db, err := sql.Open("sqlite3", opts.DatabasePath())
	if err != nil {
		return err
	}
	defer db.Close()
	sqlStmt := `
		CREATE TABLE searchIndex(id INTEGER PRIMARY KEY, name TEXT, type TEXT, path TEXT);
		CREATE UNIQUE INDEX anchor ON searchIndex (name, type, path);
		`
	if _, err = db.Exec(sqlStmt); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO searchIndex(name, type, path) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if err = filepath.Walk(opts.ContentPath(), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".htm") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			content := string(b)
			res := titleRE.FindAllStringSubmatch(content, -1)
			if len(res) >= 1 && len(res[0]) >= 2 {
				_, err := stmt.Exec(res[0][1], "Guide", strings.TrimPrefix(path, opts.ContentPath()))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func main() {
	opts := NewOptions()
	opts.Clean()
	failOnError(opts.ExtractSource())
	failOnError(opts.CreateDatabase())
	failOnError(opts.WritePlist())
}
