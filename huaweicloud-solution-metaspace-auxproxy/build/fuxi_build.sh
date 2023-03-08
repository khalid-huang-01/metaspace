#!/bin/bash

cd build
sh -x build.sh

if [ $? != 0 ]
then
 echo "make failed"
 exit -1
fi

