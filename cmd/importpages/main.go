package main

import (
	"bufio"
	"compress/bzip2"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type doc struct {
	Title string `xml:"title"`
	Text  string `xml:"revision>text"`
}

var (
	rxSpecial = regexp.MustCompile(`{{(?s).*?}}|{\|(?s).*?\|}`)
	rxLink    = regexp.MustCompile(`\[\[.+?\]\]`)
	rxTag     = regexp.MustCompile(`<[^>]+>`)
)

func (d doc) cleanText() string {
	text := d.Text
	text = rxSpecial.ReplaceAllString(text, "")
	text = rxLink.ReplaceAllStringFunc(text, func(t string) string {
		p := strings.Split(t[len("[["):len(t)-len("]]")], "|")
		return p[len(p)-1]
	})
	text = rxTag.ReplaceAllString(text, "")
	text = strings.NewReplacer("'''", "'", "&amp;", "&").Replace(text)
	return text
}

const totalCount = 2082845 // grep '<page>' | wc -l

func main() {
	log.SetPrefix(filepath.Base(os.Args[0]) + ": ")
	log.SetFlags(0)

	file := os.Args[1]
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	var iptr importer
	target := os.Args[2]
	if strings.Contains(target, `@/`) {
		db, err := sql.Open("mysql", os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		iptr = &dbImporter{db: db}
	} else {
		f, err := os.Create(target)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if strings.HasSuffix(target, ".json") {
			bio := bufio.NewWriter(f)
			iptr = &jsonImporter{w: bio}
		} else {
			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = '\t'
			iptr = &csvImporter{csv: csvWriter}
		}
	}

	var rdr io.Reader = f
	if strings.HasSuffix(file, ".bz2") {
		rdr = bzip2.NewReader(f)
	}
	decoder := xml.NewDecoder(rdr)

	iptr.Prepare()

	count := 0
	start := time.Now()
	go func() {
		tick := time.Tick(time.Second)
		for t := range tick {
			if count == 0 {
				continue
			}
			elapsed := t.Sub(start)
			eta := time.Duration(float64(elapsed/time.Second)*(float64(totalCount-count)/float64(count))) * time.Second
			log.Printf("%d/%d (%.1f%%) speed: %.1f/s ETA:%s", count, totalCount, float64(count)/totalCount*100, float64(count)/float64(elapsed/time.Second), eta)
		}
	}()

	for {
		tok, err := decoder.Token()
		if err != io.EOF && err != nil {
			log.Fatal(err)
		}
		if err == io.EOF || tok == nil {
			iptr.Finalize()
			break
		}

		if se, ok := tok.(xml.StartElement); ok {
			if se.Name.Local == "page" {
				var d doc
				err := decoder.DecodeElement(&d, &se)
				if err != nil {
					log.Fatal(err)
				}
				iptr.Import(d.Title, d.cleanText())
				count++
			}
		}
	}
}

type importer interface {
	Prepare() error
	Import(key, value string) error
	Finalize() error
}

type dbImporter struct {
	db   *sql.DB
	pool []interface{}
}

const bulkUnit = 10000

func (d *dbImporter) Prepare() error {
	d.pool = make([]interface{}, 0, bulkUnit)
	_, err := d.db.Exec("TRUNCATE TABLE entry")
	return err
}

func (d *dbImporter) Import(k, v string) error {
	d.pool = append(d.pool, k, v)
	if len(d.pool) >= bulkUnit {
		err := d.insert()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbImporter) insert() error {
	l := len(d.pool)
	if l == 0 {
		return nil
	}
	sql := "INSERT IGNORE INTO entry (keyword, description) VALUES "
	trail := strings.Repeat("(?,?),", l/2)
	trail = trail[:(len(trail) - 1)]
	sql += trail
	_, err := d.db.Exec(sql, d.pool...)
	if err != nil {
		return err
	}
	d.pool = make([]interface{}, 0, bulkUnit)
	return nil
}

func (d *dbImporter) Finalize() error {
	return d.insert()
}

type csvImporter struct {
	csv *csv.Writer
}

func (c *csvImporter) Prepare() error {
	return nil
}

func (c *csvImporter) Import(k, v string) error {
	record := []string{k, v}
	c.csv.Write(record)
	return c.csv.Error()
}

func (c *csvImporter) Finalize() error {
	c.csv.Flush()
	return c.csv.Error()
}

type jsonImporter struct {
	w *bufio.Writer
}

func (j *jsonImporter) Prepare() error {
	return nil
}

func (j *jsonImporter) Import(k, v string) error {
	jsn, err := json.Marshal(map[string]string{
		"k": k,
		"v": v,
	})
	if err != nil {
		return err
	}
	jsn = append(jsn, '\n')
	_, err = j.w.Write(jsn)
	return err
}

func (j *jsonImporter) Finalize() error {
	return j.w.Flush()
}
