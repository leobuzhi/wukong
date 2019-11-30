package core

import (
	"fmt"

	"github.com/leobuzhi/wukong/types"
)

func indicesToString(indexer *Indexer, token string) (output string) {
	if indices, ok := indexer.tableLock.table[token]; ok {
		for i := 0; i < indexer.getIndexLength(indices); i++ {
			output += fmt.Sprintf("%d ",
				indexer.getDocID(indices, i))
		}
	}
	return
}

func indexedDocsToString(docs []types.IndexedDocument, numDocs int) (output string) {
	for _, doc := range docs {
		output += fmt.Sprintf("[%d %d %v] ",
			doc.DocID, doc.TokenProximity, doc.TokenSnippetLocations)
	}
	return
}

func scoredDocsToString(docs []types.ScoredDocument) (output string) {
	for _, doc := range docs {
		output += fmt.Sprintf("[%d [", doc.DocID)
		for _, score := range doc.Scores {
			output += fmt.Sprintf("%d ", int(score*1000))
		}
		output += "]] "
	}
	return
}
