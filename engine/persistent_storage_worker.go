package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"sync/atomic"

	"github.com/leobuzhi/wukong/types"
)

type persistentStorageIndexDocumentRequest struct {
	docID uint64
	data  types.DocumentIndexData
}

func (engine *Engine) persistentStorageIndexDocumentWorker(shard int) {
	for {
		request := <-engine.persistentStorageIndexDocumentChannels[shard]

		// 得到key
		b := make([]byte, 10)
		length := binary.PutUvarint(b, request.docID)

		// 得到value
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(request.data)
		if err != nil {
			atomic.AddUint64(&engine.numDocumentsStored, 1)
			continue
		}

		// 将key-value写入数据库
		engine.dbs[shard].Set(b[0:length], buf.Bytes())
		atomic.AddUint64(&engine.numDocumentsStored, 1)
	}
}

func (engine *Engine) persistentStorageRemoveDocumentWorker(docID uint64, shard uint32) {
	// 得到key
	b := make([]byte, 10)
	length := binary.PutUvarint(b, docID)

	// 从数据库删除该key
	engine.dbs[shard].Delete(b[0:length])
}

func (engine *Engine) persistentStorageInitWorker(shard int) {
	engine.dbs[shard].ForEach(func(k, v []byte) error {
		key, value := k, v
		// 得到docID
		docID, _ := binary.Uvarint(key)

		// 得到data
		buf := bytes.NewReader(value)
		dec := gob.NewDecoder(buf)
		var data types.DocumentIndexData
		err := dec.Decode(&data)
		if err == nil {
			// 添加索引
			engine.internalIndexDocument(docID, data, false)
		}
		return nil
	})
	engine.persistentStorageInitChannel <- true
}
