// Copyright 2018 The CubeFS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package master

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cubefs/cubefs/proto"
)

type Decoder func([]byte) ([]byte, error)

func (d Decoder) Decode(raw []byte) ([]byte, error) {
	return d(raw)
}

type ClientAPI struct {
	mc *MasterClient
	h  map[string]string // extra headers
}

func (api *ClientAPI) WithHeader(key, val string) *ClientAPI {
	return &ClientAPI{mc: api.mc, h: mergeHeader(api.h, key, val)}
}

func (api *ClientAPI) EncodingWith(encoding string) *ClientAPI {
	return api.WithHeader(headerAcceptEncoding, encoding)
}

func (api *ClientAPI) EncodingGzip() *ClientAPI {
	return api.EncodingWith(encodingGzip)
}

func (api *ClientAPI) GetVolume(ctx context.Context, volName string, authKey string) (vv *proto.VolView, err error) {
	vv = &proto.VolView{}
	ctx = proto.ContextWithOperation(ctx, "GetVolume")
	err = api.mc.requestWith(vv, newRequest(ctx, post, proto.ClientVol).
		Header(api.h).Param(anyParam{"name", volName}, anyParam{"authKey", authKey}))
	return
}

func (api *ClientAPI) GetVolumeWithoutAuthKey(ctx context.Context, volName string) (vv *proto.VolView, err error) {
	vv = &proto.VolView{}
	ctx = proto.ContextWithOperation(ctx, "GetVolumeWithoutAuthKey")
	err = api.mc.requestWith(vv, newRequest(ctx, post, proto.ClientVol).
		Header(api.h, proto.SkipOwnerValidation, "true").addParam("name", volName))
	return
}

func (api *ClientAPI) GetVolumeWithAuthnode(ctx context.Context, volName string, authKey string, token string, decoder Decoder) (vv *proto.VolView, err error) {
	var body []byte
	ctx = proto.ContextWithOperation(ctx, "GetVolumeWithAuthnode")
	request := newRequest(ctx, post, proto.ClientVol).Header(api.h)
	request.addParam("name", volName)
	request.addParam("authKey", authKey)
	request.addParam(proto.ClientMessage, token)
	if body, err = api.mc.serveRequest(request); err != nil {
		return
	}
	if decoder != nil {
		if body, err = decoder.Decode(body); err != nil {
			return
		}
	}
	vv = &proto.VolView{}
	if err = json.Unmarshal(body, vv); err != nil {
		return
	}
	return
}

func (api *ClientAPI) GetVolumeStat(ctx context.Context, volName string) (info *proto.VolStatInfo, err error) {
	info = &proto.VolStatInfo{}
	ctx = proto.ContextWithOperation(ctx, "GetVolumeStat")
	err = api.mc.requestWith(info, newRequest(ctx, get, proto.ClientVolStat).
		Header(api.h).Param(anyParam{"name", volName}, anyParam{"version", proto.LFClient}))
	return
}

func (api *ClientAPI) GetMetaPartition(ctx context.Context, partitionID uint64) (partition *proto.MetaPartitionInfo, err error) {
	partition = &proto.MetaPartitionInfo{}
	ctx = proto.ContextWithOperation(ctx, "GetMetaPartition")
	err = api.mc.requestWith(partition, newRequest(ctx, get, proto.ClientMetaPartition).
		Header(api.h).addParamAny("id", partitionID))
	return
}

func (api *ClientAPI) GetMetaPartitions(ctx context.Context, volName string) (views []*proto.MetaPartitionView, err error) {
	views = make([]*proto.MetaPartitionView, 0)
	ctx = proto.ContextWithOperation(ctx, "GetMetaPartitions")
	err = api.mc.requestWith(&views, newRequest(ctx, get, proto.ClientMetaPartitions).
		Header(api.h).addParam("name", volName))
	return
}

func (api *ClientAPI) GetDataPartitions(ctx context.Context, volName string) (view *proto.DataPartitionsView, err error) {
	ctx = proto.ContextWithOperation(ctx, "GetDataPartitions")
	request := newRequest(ctx, get, proto.ClientDataPartitions).Header(api.h).addParam("name", volName)

	lastLeader := api.mc.leaderAddr
	defer api.mc.SetLeader(lastLeader)
	randIndex := rand.Intn(len(api.mc.masters))
	if randIndex >= len(api.mc.masters) {
		err = fmt.Errorf("master len %v less or equal request index %v", len(api.mc.masters), randIndex)
		return
	}
	api.mc.SetLeader(api.mc.masters[randIndex])
	var data []byte
	if data, err = api.mc.serveRequest(request); err != nil {
		return
	}
	view = &proto.DataPartitionsView{}
	if err = json.Unmarshal(data, view); err != nil {
		return
	}
	return
}

func (api *ClientAPI) GetPreLoadDataPartitions(ctx context.Context, volName string) (view *proto.DataPartitionsView, err error) {
	view = &proto.DataPartitionsView{}
	ctx = proto.ContextWithOperation(ctx, "GetPreLoadDataPartitions")
	err = api.mc.requestWith(view, newRequest(ctx, get, proto.ClientDataPartitions).
		Header(api.h).addParam("name", volName))
	return
}
