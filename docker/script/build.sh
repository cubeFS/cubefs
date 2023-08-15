#!/usr/bin/env bash
mkdir -p /go/src/github.com/cubefs/cubefs/docker/bin;
failed=0

build_opt=
case $1 in
	test)
		echo Build mode: TESING
		build_opt='test'
		;;

	*)
		echo Build mode: NORMAL	
esac

cd /go/src/github.com/cubefs/cubefs
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
echo "Branch: ${BranchName}"
echo "Commit: ${CommitID}"

echo -n 'Building ChubaoFS Server ... ';
cd /go/src/github.com/cubefs/cubefs/cmd;
bash ./build.sh ${build_opt} &>> /tmp/cfs_build_output
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mdone\033[0m";
    mv cfs-server /go/src/github.com/cubefs/cubefs/docker/bin/cfs-server;
else
    echo -e "\033[31mfail\033[0m";
    failed=1
fi


echo -n 'Building ChubaoFS Client ... ' ;
cd /go/src/github.com/cubefs/cubefs/client;
bash ./build.sh -d ${build_opt} &>> /tmp/cfs_build_output
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mdone\033[0m";
    mv bin/cfs-client /go/src/github.com/cubefs/cubefs/docker/bin/cfs-client;
    mv bin/cfs-client-inner /go/src/github.com/cubefs/cubefs/docker/bin/cfs-client-inner;
    mv bin/libcfssdk.so /go/src/github.com/cubefs/cubefs/docker/bin/libcfssdk.so;
    mv bin/libcfsc.so /go/src/github.com/cubefs/cubefs/docker/bin/libcfsc.so;
    mv bin/libcfssdk_cshared.so /go/src/github.com/cubefs/cubefs/docker/bin/libcfssdk_cshared.so;
    mv bin/libcfsclient.so /go/src/github.com/cubefs/cubefs/docker/bin/libcfsclient.so;
    mv bin/libempty.so /go/src/github.com/cubefs/cubefs/docker/bin/libempty.so;
    mv /usr/local/go/pkg/linux_amd64_dynlink/libstd.so /go/src/github.com/cubefs/cubefs/docker/bin/libstd.so;
    if [ "${build_opt}"x = "test"x ]; then
        mv bin/test-bypass /go/src/github.com/cubefs/cubefs/docker/bin/test-bypass;
    fi
else
    echo -e "\033[31mfail\033[0m";
    failed=1
fi

echo -n 'Building ChubaoFS CLI    ... ';
cd /go/src/github.com/cubefs/cubefs/cli;
bash ./build.sh ${build_opt} &>> /tmp/cfs_build_output;
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mdone\033[0m";
    mv cfs-cli /go/src/github.com/cubefs/cubefs/docker/bin/cfs-cli;
else
    echo -e "\033[31mfail\033[0m";
    failed=1
fi


echo -n 'Building ChubaoFS repair server    ... ';
cd /go/src/github.com/cubefs/cubefs/cli/repaircrc;
bash ./build.sh ${build_opt} &>> /tmp/cfs_build_output;
if [[ $? -eq 0 ]]; then
    echo -e "\033[32mdone\033[0m";
    mv repair_server /go/src/github.com/cubefs/cubefs/docker/bin/repair_server;
else
    echo -e "\033[31mfail\033[0m";
    failed=1
fi

if [[ ${failed} -eq 1 ]]; then
    echo -e "\nbuild output:"
    cat /tmp/cfs_build_output;
    exit 1
fi
cat /tmp/cfs_build_output;
exit 0
