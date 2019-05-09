// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"net/http"
	"sync"
)

type pipeline interface {
	run(string, string)
}

//PipelineConfig contains main options of pipeline
type PipelineConfig struct {
	Ctx                    context.Context
	ParsingProcessesCount  int
	CountingProcessesCount int
	Timeout                int
}

type htmlParams struct {
	ctx           context.Context
	client        *http.Client
	timeout       int
	maxGoroutines int
}

type resultStruct struct {
	url   string
	count int
}

type parserStruct struct {
	url  string
	html string
}

func runPipeline(pipeline pipeline, urls, searchString string) {
	pipeline.run(urls, searchString)
}

//Parse starting pipeline
func Parse(config PipelineConfig, wg *sync.WaitGroup, urls, searchString string) {
	pipelinePr := pipelinePreload{config: config}
	runPipeline(pipelinePr, urls, searchString)

	//pipelineCommon := pipelineCommon{config: config}
	//runPipeline(pipelineCommon, urls, searchString)
	wg.Done()
}
