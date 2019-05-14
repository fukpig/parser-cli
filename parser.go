// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"sync"
)

//nolint:unused
type pipeline interface {
	run(context.Context)
}

//PipelineConfig contains main options of pipeline
type PipelineConfig struct {
	ParsingProcessesCount  int
	CountingProcessesCount int
	Timeout                int
	PipelineType           string
	Wg                     *sync.WaitGroup
	Urls                   string
	SearchString           string
}

type htmlParams struct {
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

//Parse starting pipeline
func Parse(ctx context.Context, config PipelineConfig) {

	var pipe pipeline
	if config.PipelineType == "preload" {
		pipe = pipelinePreload{config: config}
	} else {
		pipe = pipelineCommon{config: config}
	}
	pipe.run(ctx)
	config.Wg.Done()
}
