#!/bin/bash

root_dir="/home/scase/bin"
service_name="AASS"
report_file_dir="/home/scase/goc"

while true
do
  #每60秒获取一次
  sleep 10m
  #收集覆盖率写入 LiveOriginService.txt文件
  mkdir ${report_file_dir}
  ${root_dir}/goc profile  > ${report_file_dir}/${service_name}.out && ${root_dir}/go/bin/go tool cover -func=${report_file_dir}/${service_name}.out > ${report_file_dir}/${service_name}.txt
  ${root_dir}/go/bin/go tool cover -html=${report_file_dir}/${service_name}.out -o ${report_file_dir}/${service_name}.html
  curl --form "fileUpload=@${report_file_dir}/${service_name}.html" http://${TEST_SERVER}/goc/cover/save/file/${service_name}/${TEST_CLUSTER}
  curl --form "fileUpload=@${report_file_dir}/${service_name}.out" http://${TEST_SERVER}/goc/cover/upload/cloudTest/${service_name}/${TEST_CLUSTER}
done
