package engine

import (
	"sync/atomic"

	"github.com/leobuzhi/wukong/types"
)

type indexerAddDocumentRequest struct {
	document    *types.DocumentIndex
	forceUpdate bool
}

type indexerLookupRequest struct {
	countDocsOnly       bool
	tokens              []string
	labels              []string
	docIDs              map[uint64]bool
	options             types.RankOptions
	rankerReturnChannel chan rankerReturnRequest
	orderless           bool
}

type indexerRemoveDocRequest struct {
	docID       uint64
	forceUpdate bool
}

func (engine *Engine) indexerAddDocumentWorker(shard int) {
	for {
		request := <-engine.indexerAddDocChannels[shard]
		engine.indexers[shard].AddDocumentToCache(request.document, request.forceUpdate)
		if request.document != nil {
			atomic.AddUint64(&engine.numTokenIndexAdded,
				uint64(len(request.document.Keywords)))
			atomic.AddUint64(&engine.numDocumentsIndexed, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numDocumentsForceUpdated, 1)
		}
	}
}

func (engine *Engine) indexerRemoveDocWorker(shard int) {
	for {
		request := <-engine.indexerRemoveDocChannels[shard]
		engine.indexers[shard].RemoveDocumentToCache(request.docID, request.forceUpdate)
		if request.docID != 0 {
			atomic.AddUint64(&engine.numDocumentsRemoved, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numDocumentsForceUpdated, 1)
		}
	}
}

func (engine *Engine) indexerLookupWorker(shard int) {
	for {
		request := <-engine.indexerLookupChannels[shard]

		var docs []types.IndexedDocument
		var numDocs int
		if request.docIDs == nil {
			docs, numDocs = engine.indexers[shard].Lookup(request.tokens, request.labels, nil, request.countDocsOnly)
		} else {
			docs, numDocs = engine.indexers[shard].Lookup(request.tokens, request.labels, request.docIDs, request.countDocsOnly)
		}

		if request.countDocsOnly {
			request.rankerReturnChannel <- rankerReturnRequest{numDocs: numDocs}
			continue
		}

		if len(docs) == 0 {
			request.rankerReturnChannel <- rankerReturnRequest{}
			continue
		}

		if request.orderless {
			var outputDocs []types.ScoredDocument
			for _, d := range docs {
				outputDocs = append(outputDocs, types.ScoredDocument{
					DocID: d.DocID,
					TokenSnippetLocations: d.TokenSnippetLocations,
					TokenLocations:        d.TokenLocations})
			}
			request.rankerReturnChannel <- rankerReturnRequest{
				docs:    outputDocs,
				numDocs: len(outputDocs),
			}
			continue
		}

		rankerRequest := rankerRankRequest{
			countDocsOnly:       request.countDocsOnly,
			docs:                docs,
			options:             request.options,
			rankerReturnChannel: request.rankerReturnChannel,
		}
		engine.rankerRankChannels[shard] <- rankerRequest
	}
}
