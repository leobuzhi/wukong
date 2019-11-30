package core

import (
	"log"
	"sort"
	"sync"

	"github.com/leobuzhi/wukong/types"
	"github.com/leobuzhi/wukong/utils"
)

type Ranker struct {
	lock struct {
		sync.RWMutex
		fields map[uint64]interface{}
		docs   map[uint64]bool
	}
	initialized bool
}

func (ranker *Ranker) Init() {
	if ranker.initialized == true {
		log.Fatal("ranker init twice")
	}
	ranker.initialized = true

	ranker.lock.fields = make(map[uint64]interface{})
	ranker.lock.docs = make(map[uint64]bool)
}

// 给某个文档添加评分字段
func (ranker *Ranker) AddDoc(docID uint64, fields interface{}) {
	if ranker.initialized == false {
		log.Fatal("The ranker has not been initialized")
	}

	ranker.lock.Lock()
	ranker.lock.fields[docID] = fields
	ranker.lock.docs[docID] = true
	ranker.lock.Unlock()
}

// 删除某个文档的评分字段
func (ranker *Ranker) RemoveDoc(docID uint64) {
	if ranker.initialized == false {
		log.Fatal("The ranker has not been initialized")
	}

	ranker.lock.Lock()
	delete(ranker.lock.fields, docID)
	delete(ranker.lock.docs, docID)
	ranker.lock.Unlock()
}

// 给文档评分并排序
func (ranker *Ranker) Rank(
	docs []types.IndexedDocument, options types.RankOptions, countDocsOnly bool) (types.ScoredDocuments, int) {
	if ranker.initialized == false {
		log.Fatal("The ranker has not been initialized")
	}

	// 对每个文档评分
	var outputDocs types.ScoredDocuments
	numDocs := 0
	for _, d := range docs {
		ranker.lock.RLock()
		// 判断 doc 是否存在
		if _, ok := ranker.lock.docs[d.DocID]; ok {
			fs := ranker.lock.fields[d.DocID]
			ranker.lock.RUnlock()
			// 计算评分并剔除没有分值的文档
			scores := options.ScoringCriteria.Score(d, fs)
			if len(scores) > 0 {
				if !countDocsOnly {
					outputDocs = append(outputDocs, types.ScoredDocument{
						DocID:                 d.DocID,
						Scores:                scores,
						TokenSnippetLocations: d.TokenSnippetLocations,
						TokenLocations:        d.TokenLocations})
				}
				numDocs++
			}
		} else {
			ranker.lock.RUnlock()
		}
	}

	// 排序
	if !countDocsOnly {
		if options.ReverseOrder {
			sort.Sort(sort.Reverse(outputDocs))
		} else {
			sort.Sort(outputDocs)
		}
		// 当用户要求只返回部分结果时返回部分结果
		var start, end int
		if options.MaxOutputs != 0 {
			start = utils.MinInt(options.OutputOffset, len(outputDocs))
			end = utils.MinInt(options.OutputOffset+options.MaxOutputs, len(outputDocs))
		} else {
			start = utils.MinInt(options.OutputOffset, len(outputDocs))
			end = len(outputDocs)
		}
		return outputDocs[start:end], numDocs
	}
	return outputDocs, numDocs
}
