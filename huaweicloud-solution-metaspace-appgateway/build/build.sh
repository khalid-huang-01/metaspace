#!/bin/bash

cur_dir=`pwd`
root_dir=${cur_dir}/../
export GO111MODULE=on
export GOPROXY=https://cmc.centralrepo.rnd.huawei.com/cbu-go,direct
export GONOSUMDB=*
export GOSUMDB=off

artifactTag=""

function init(){
    YYYY=`echo ${CID_BUILD_TIME:0:4}`
    MMdd=`echo ${CID_BUILD_TIME:4:4}`
    hhmm=`echo ${CID_BUILD_TIME:8:4}`
    ss=`echo ${CID_BUILD_TIME:12:2}`
    artifactTag="${CID_REPO_COMMIT}_${YYYY}.${MMdd}.${hhmm}.${ss}"
    bm -action download -config bm.json
    rpm -Uvh packages/seccomponent-1.1.5-release.x86_64.rpm
    ln -s /usr/local/seccomponent/lib/go/src/cryptoapi ${GOROOT}/src/cryptoapi
}


function build_pro() {
  cd ${root_dir}
  make
  if [ $? != 0 ]
  then
   echo "make failed"
   exit -1
  fi
}

# 云视频静态检查要求支持上报代码覆盖率
function build_goc() {
  echo "build for golang coverage"

  sed -i 's/TEST_SERVER/'${TEST_SERVER}'/' ${root_dir}/build/collect_coverage.sh
  sed -i 's/TEST_CLUSTER/'${TEST_CLUSTER}'/' ${root_dir}/build/collect_coverage.sh

  bm -action download -name goc-server -version 1.3.3 -file goc -release latest -output ./
  bm -action download -name go-sdk -version 1.13 -file go1.13.linux-amd64.tar.gz -release latest -output ./
  chmod 750 ./goc

  ./goc build -o ${root_dir}/bin/appgateway . --debug
  ls -l
  cp ./goc ${root_dir}/bin/goc
  mkdir ${root_dir}/bin/src
  cp ${root_dir} ${root_dir}/bin/src -rf
  cp ${root_dir}/build/collect_coverage.sh ${root_dir}/bin/
  cp ./go1.13.linux-amd64.tar.gz ${root_dir}/bin/go1.13.linux-amd64.tar.gz
  cd ../
}

function build() {
  init
  if [ ${GOC_VALUE} ] && [ ${GOC_VALUE} == true ]
  then
      build_goc
  else
      build_pro
  fi
}

build