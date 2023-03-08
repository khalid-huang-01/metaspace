#!/bin/bash

cd build
sh -x build.sh

if [ $? != 0 ]
then
 echo "make failed"
 exit -1
fi

cd -
YYYY=`echo ${CID_BUILD_TIME:0:4}`
MMdd=`echo ${CID_BUILD_TIME:4:4}`
hhmm=`echo ${CID_BUILD_TIME:8:4}`
ss=`echo ${CID_BUILD_TIME:12:2}`
docker build --no-cache -t registry-cbu.huawei.com/sparknext/aass:${CID_REPO_COMMIT}_${YYYY}.${MMdd}.${hhmm}.${ss} ./

echo registry-cbu.huawei.com/sparknext/aass:${CID_REPO_COMMIT}_${YYYY}.${MMdd}.${hhmm}.${ss} > version.txt
