#!/usr/bin/env bash
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
BuildTime=`date +%Y-%m-%d\ %H:%M`

build_opt=$1

[[ "-$GOPATH" == "-" ]] && { echo "GOPATH not set"; exit 1; }

case ${build_opt} in
	test)
		go test -c -covermode=atomic -coverpkg="../..." -ldflags "-X main.CommitID=${CommitID} -X main.BranchName=${BranchName} -X 'main.BuildTime=${BuildTime}' -X github.com/cubefs/cubefs/proto.ENV=test" -o cfs-server
		;;
  all-test)
    go build -ldflags "-X main.CommitID=${CommitID} -X main.BranchName=${BranchName} -X 'main.BuildTime=${BuildTime}' -X github.com/cubefs/cubefs/proto.ENV=test" -o cfs-server
    ;;
	*)
    go build -ldflags "-X main.CommitID=${CommitID} -X main.BranchName=${BranchName} -X 'main.BuildTime=${BuildTime}'" -o cfs-server
esac


