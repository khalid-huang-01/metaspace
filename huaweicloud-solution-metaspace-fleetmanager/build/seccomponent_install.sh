#!/bin/bash

cur_dir=`pwd`
root_dir=${cur_dir}/../
export GO111MODULE=on
export GOPROXY=https://cmc.centralrepo.rnd.huawei.com/cbu-go,direct
export GONOSUMDB=*
export GOSUMDB=off

function install(){
    bm -action download -config bm.json
    rpm -Uvh packages/seccomponent-1.1.5-release.x86_64.rpm
    ln -s /usr/local/seccomponent/lib/go/src/cryptoapi ${GOROOT}/src/cryptoapi
}

install
