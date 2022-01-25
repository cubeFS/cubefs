package blobstore

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/chubaofs/chubaofs/util/stat"

	"github.com/chubaofs/chubaofs/blockcache/bcache"
	"github.com/chubaofs/chubaofs/proto"
	"github.com/chubaofs/chubaofs/sdk/data/stream"
	"github.com/chubaofs/chubaofs/sdk/meta"
	"github.com/chubaofs/chubaofs/util"
	"github.com/chubaofs/chubaofs/util/log"
)

const (
	MaxBufferSize = 512 * util.MB
)

type Writer struct {
	volType      int
	volName      string
	blockSize    int
	ino          uint64
	err          chan error
	bc           *bcache.BcacheClient
	mw           *meta.MetaWrapper
	ec           *stream.ExtentClient
	ebsc         *BlobStoreClient
	wConcurrency int
	wg           sync.WaitGroup
	once         sync.Once
	sync.Mutex
	enableBcache   bool
	cacheAction    int
	buf            []byte
	fileOffset     int
	fileCache      bool
	fileSize       uint64
	cacheThreshold int
	dirty          bool
}

func NewWriter(config ClientConfig) (writer *Writer) {
	writer = new(Writer)

	writer.volName = config.VolName
	writer.volType = config.VolType
	writer.blockSize = config.BlockSize
	writer.ino = config.Ino
	writer.err = nil
	writer.bc = config.Bc
	writer.mw = config.Mw
	writer.ec = config.Ec
	writer.ebsc = config.Ebsc
	writer.wConcurrency = config.WConcurrency
	writer.wg = sync.WaitGroup{}
	writer.once = sync.Once{}
	writer.Mutex = sync.Mutex{}
	writer.enableBcache = config.EnableBcache
	writer.cacheAction = config.CacheAction
	writer.fileCache = config.FileCache
	writer.fileSize = config.FileSize
	writer.cacheThreshold = config.CacheThreshold
	writer.dirty = false

	return
}

func (writer *Writer) String() string {
	return fmt.Sprintf("Writer{address(%v),volName(%v),volType(%v),ino(%v),blockSize(%v),fileSize(%v),enableBcache(%v),cacheAction(%v),fileCache(%v),cacheThreshold(%v)},wConcurrency(%v)",
		&writer, writer.volName, writer.volType, writer.ino, writer.blockSize, writer.fileSize, writer.enableBcache, writer.cacheAction, writer.fileCache, writer.cacheThreshold, writer.wConcurrency)
}

func (writer *Writer) Write(ctx context.Context, offset int, data []byte, flags int) (size int, err error) {
	//atomic.StoreInt32(&writer.idle, 0)
	if writer == nil {
		return 0, fmt.Errorf("writer is not opened yet")
	}
	log.LogDebugf("TRACE blobStore Write Enter: ino(%v) offset(%v) len(%v) flags&proto.FlagsAppend(%v) fileSize(%v)", writer.ino, offset, len(data), flags&proto.FlagsAppend, writer.CacheFileSize())

	if len(data) > MaxBufferSize || flags&proto.FlagsAppend == 0 || offset != writer.CacheFileSize() {
		log.LogErrorf("TRACE blobStore Write error,may be len(%v)>512MB,flags(%v)!=flagAppend,offset(%v)!=fileSize(%v)", len(data), flags&proto.FlagsAppend, offset, writer.CacheFileSize())
		err = syscall.EOPNOTSUPP
		return
	}
	//write buffer
	log.LogDebugf("TRACE blobStore Write: ino(%v) offset(%v) len(%v) flags&proto.FlagsSyncWrite(%v)", writer.ino, offset, len(data), flags&proto.FlagsSyncWrite)
	if flags&proto.FlagsSyncWrite == 0 {
		size, err = writer.doBufferWrite(ctx, data, offset)
		return
	}
	//parallel io write ebs direct
	size, err = writer.doParallelWrite(ctx, data, offset)
	return
}

func (writer *Writer) doParallelWrite(ctx context.Context, data []byte, offset int) (size int, err error) {
	log.LogDebugf("TRACE blobStore doDirectWrite: ino(%v) offset(%v) len(%v)", writer.ino, offset, len(data))
	writer.Lock()
	defer writer.Unlock()
	wSlices := writer.prepareWriteSlice(offset, data)
	log.LogDebugf("TRACE blobStore prepareWriteSlice: wSlices(%v)", wSlices)
	sliceSize := len(wSlices)

	writer.wg.Add(sliceSize)
	writer.err = make(chan error, sliceSize)
	pool := New(writer.wConcurrency, sliceSize)
	defer pool.Close()
	for _, wSlice := range wSlices {
		pool.Execute(wSlice, func(param *rwSlice) {
			writer.writeSlice(ctx, param, true)
		})
	}
	writer.wg.Wait()
	for i := 0; i < sliceSize; i++ {
		if err, ok := <-writer.err; !ok || err != nil {
			log.LogErrorf("slice write error,ino(%v) fileoffset(%v) sliceOffset(%v) sliceSize(%v) err(%v)", writer.ino, wSlices[i].fileOffset, wSlices[i].rOffset, wSlices[i].rSize, err)
			return 0, err
		}
	}
	close(writer.err)
	//update meta
	oeks := make([]proto.ObjExtentKey, 0)
	for _, wSlice := range wSlices {
		size += int(wSlice.size)
		oeks = append(oeks, wSlice.objExtentKey)
	}
	log.LogDebugf("TRACE blobStore appendObjExtentKeys: oeks(%v)", oeks)
	if err = writer.mw.AppendObjExtentKeys(writer.ino, oeks); err != nil {
		log.LogErrorf("slice write error,meta append ebsc extent keys fail,ino(%v) fileOffset(%v) len(%v) err(%v)", writer.ino, offset, len(data), err)
		return
	}
	atomic.AddUint64(&writer.fileSize, uint64(size))

	for _, wSlice := range wSlices {
		writer.cacheLevel2(wSlice)
	}

	return
}

func (writer *Writer) cacheLevel2(wSlice *rwSlice) {
	if writer.cacheAction == proto.RWCache && (wSlice.fileOffset+uint64(wSlice.size)) < uint64(writer.cacheThreshold) || writer.fileCache {
		buf := make([]byte, wSlice.size)
		offSet := int(wSlice.fileOffset)
		copy(buf, wSlice.Data)
		go writer.asyncCache(writer.ino, offSet, buf)
	}
}

func (writer *Writer) doBufferWrite(ctx context.Context, data []byte, offset int) (size int, err error) {
	log.LogDebugf("TRACE blobStore doBufferWrite Enter: ino(%v) offset(%v) len(%v)", writer.ino, offset, len(data))

	writer.fileOffset = offset
	dataSize := len(data)
	position := 0
	log.LogDebugf("TRACE blobStore doBufferWrite: ino(%v) writer.buf.len(%v) writer.blocksize(%v)", writer.ino, len(writer.buf), writer.blockSize)
	writer.Lock()
	defer writer.Unlock()
	for dataSize > 0 {
		freeSize := writer.blockSize - len(writer.buf)
		if dataSize < freeSize {
			freeSize = dataSize
		}
		log.LogDebugf("TRACE blobStore doBufferWrite: ino(%v) writer.fileSize(%v) writer.fileOffset(%v) position(%v) freeSize(%v)", writer.ino, writer.fileSize, writer.fileOffset, position, freeSize)
		writer.buf = append(writer.buf, data[position:position+freeSize]...)
		log.LogDebugf("TRACE blobStore doBufferWrite:ino(%v) writer.buf.len(%v)", writer.ino, len(writer.buf))
		position += freeSize
		dataSize -= freeSize
		writer.fileOffset += freeSize
		writer.dirty = true

		if len(writer.buf) == writer.blockSize {
			log.LogDebugf("TRACE blobStore doBufferWrite: ino(%v) writer.buf.len(%v) writer.blocksize(%v)", writer.ino, len(writer.buf), writer.blockSize)
			writer.Unlock()
			err = writer.flush(writer.ino, ctx, false)
			writer.Lock()
			if err != nil {
				writer.buf = writer.buf[:len(writer.buf)-len(data)]
				writer.fileOffset -= len(data)
				return
			}

		}
	}

	size = len(data)
	atomic.AddUint64(&writer.fileSize, uint64(size))

	log.LogDebugf("TRACE blobStore doBufferWrite Exit: ino(%v) writer.fileSize(%v) writer.fileOffset(%v)", writer.ino, writer.fileSize, writer.fileOffset)
	return size, nil
}

func (writer *Writer) Flush(ino uint64, ctx context.Context) (err error) {
	if writer == nil {
		return
	}
	if writer.shouldCacheCfs() {
		writer.ec.Flush(ino)
	}
	return writer.flush(ino, ctx, true)
}

func (writer *Writer) shouldCacheCfs() bool {
	return writer.cacheAction == proto.RWCache
}

func (writer *Writer) prepareWriteSlice(offset int, data []byte) []*rwSlice {
	size := len(data)
	wSlices := make([]*rwSlice, 0)
	wSliceCount := size / writer.blockSize
	remainSize := size % writer.blockSize
	for index := 0; index < wSliceCount; index++ {
		offset := offset + index*writer.blockSize
		wSlice := &rwSlice{
			index:      index,
			fileOffset: uint64(offset),
			size:       uint32(writer.blockSize),
			Data:       data[index*writer.blockSize : (index+1)*writer.blockSize],
		}
		wSlices = append(wSlices, wSlice)
	}
	offset = offset + wSliceCount*writer.blockSize
	if remainSize > 0 {
		wSlice := &rwSlice{
			index:      wSliceCount,
			fileOffset: uint64(offset),
			size:       uint32(remainSize),
			Data:       data[wSliceCount*writer.blockSize:],
		}
		wSlices = append(wSlices, wSlice)
	}

	return wSlices
}

func (writer *Writer) writeSlice(ctx context.Context, wSlice *rwSlice, wg bool) (err error) {
	if wg {
		defer writer.wg.Done()
	}
	log.LogDebugf("TRACE blobStore,writeSlice to ebs. ino(%v) fileOffset(%v) len(%v)", writer.ino, wSlice.fileOffset, wSlice.size)
	location, err := writer.ebsc.Write(ctx, writer.volName, wSlice.Data)
	if err != nil {
		if wg {
			writer.err <- err
		}
		return err
	}
	log.LogDebugf("TRACE blobStore,location(%v)", location)
	blobs := make([]proto.Blob, 0)
	for _, info := range location.Blobs {
		blob := proto.Blob{
			MinBid: uint64(info.MinBid),
			Count:  uint64(info.Count),
			Vid:    uint64(info.Vid),
		}
		blobs = append(blobs, blob)
	}
	wSlice.objExtentKey = proto.ObjExtentKey{
		Cid:        uint64(location.ClusterID),
		CodeMode:   uint8(location.CodeMode),
		Size:       location.Size,
		BlobSize:   location.BlobSize,
		Blobs:      blobs,
		BlobsLen:   uint32(len(blobs)),
		FileOffset: wSlice.fileOffset,
		Crc:        location.Crc,
	}
	log.LogDebugf("TRACE blobStore,objExtentKey(%v)", wSlice.objExtentKey)

	if wg {
		writer.err <- nil
	}
	return
}

func (writer *Writer) asyncCache(ino uint64, offset int, data []byte) {
	var err error
	bgTime := stat.BeginStat()
	defer func() {
		stat.EndStat("write-async-cache", err, bgTime, 1)
	}()

	log.LogDebugf("TRACE asyncCache Enter,fileOffset(%v) len(%v)", offset, len(data))
	write, err := writer.ec.Write(ino, offset, data, 0)
	log.LogDebugf("TRACE asyncCache Exit,write(%v) err(%v)", write, err)

}

func (writer *Writer) resetBuffer() {
	writer.buf = writer.buf[:0]
}

func (writer *Writer) flush(inode uint64, ctx context.Context, flushFlag bool) (err error) {
	bgTime := stat.BeginStat()
	defer func() {
		stat.EndStat("blobstore-flush", err, bgTime, 1)
	}()

	log.LogDebugf("TRACE blobStore flush: ino(%v) buf-len(%v) flushFlag(%v)", inode, len(writer.buf), flushFlag)
	writer.Lock()
	defer func() {
		writer.dirty = false
		writer.Unlock()
	}()

	if len(writer.buf) == 0 || !writer.dirty {
		return
	}
	bufferSize := len(writer.buf)
	wSlice := &rwSlice{
		fileOffset: uint64(writer.fileOffset - bufferSize),
		size:       uint32(bufferSize),
		Data:       writer.buf,
	}
	err = writer.writeSlice(ctx, wSlice, false)
	if err != nil {
		if flushFlag {
			atomic.AddUint64(&writer.fileSize, -uint64(bufferSize))
		}
		return
	}

	oeks := make([]proto.ObjExtentKey, 0)
	//update meta
	oeks = append(oeks, wSlice.objExtentKey)
	if err = writer.mw.AppendObjExtentKeys(writer.ino, oeks); err != nil {
		log.LogErrorf("slice write error,meta append ebsc extent keys fail,ino(%v) fileOffset(%v) len(%v) err(%v)", inode, wSlice.fileOffset, wSlice.size, err)
		return
	}
	writer.resetBuffer()

	writer.cacheLevel2(wSlice)
	return
}

func (writer *Writer) CacheFileSize() int {
	return int(atomic.LoadUint64(&writer.fileSize))
}
