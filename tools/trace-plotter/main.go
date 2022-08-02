// MIT License
//
// Copyright (c) 2022 Dohyun Park and EASL lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/go-echarts/go-echarts/v2/components"

	log "github.com/sirupsen/logrus"
)

var (
	elasticsearchURL = flag.String("elasticsearchURL", "http://127.0.0.1:9200", "Elasticsearch URL")
	pageSize         = flag.Int("pageSize", 100, "The number of traces to fetch per page while paginating")
	zipkinURL        = flag.String("zipkinURL", "http://127.0.0.1:8080", "Zipkin URL")
	fileName         = flag.String("fileName", "plot.html", "output file name")
	download         = flag.String("download", "", "Download all JSON trace files and save them to provided dir")
)

func main() {
	flag.Parse()
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{*elasticsearchURL},
	})
	if err != nil {
		log.Fatal(err)
	}
	l := NewLoader(es)

	var traces []Trace
	l.GetTraces(*pageSize, &traces)

	page := components.NewPage()
	page.PageTitle = "Trace Plots"

	parsedE2ETraces, e2eDurations := ParseTraces(traces, "e2e")
	parsedSystemTraces, systemDurations := ParseTraces(traces, "system")
	page.AddCharts(
		PlotGraph(parsedE2ETraces, e2eDurations, *zipkinURL, "e2e"),
		PlotGraph(parsedSystemTraces, systemDurations, *zipkinURL, "system"),
	)

	if *download != "" {
		downloadTraces(parsedE2ETraces)
	}

	if filepath.Ext(*fileName) != ".html" {
		*fileName += ".html"
	}
	f, _ := os.Create(*fileName)
	if err := page.Render(io.MultiWriter(f)); err != nil {
		log.Errorf("Error rendering plot: %s", err)
	}
}

func downloadTraces(traces []*Trace) {
	var wg sync.WaitGroup
	wg.Add(len(traces))

	if _, err := os.Stat(*download); os.IsNotExist(err) {
		err := os.MkdirAll(*download, 0700)
		if err != nil {
			panic(err)
		}
	}

	for _, trace := range traces {
		t := trace

		go func() {
			defer wg.Done()

			traceID := t.TraceID
			url := *zipkinURL + "/zipkin/api/v2/trace/" + traceID

			downloadLocation := *download + "/" + traceID + ".json"

			err := DownloadFile(downloadLocation, url)
			if err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}
