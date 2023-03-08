#!/bin/bash

install_path=$(pwd)

function start_filebeat() {
    sh ./bin/filebeat.sh
}

# 非生产环境上报代码覆盖率
function start_coverage_report() {
  if [ ${GOC_VALUE} ] && [ ${GOC_VALUE} == true ]
  then
      root_dir="/home/fleetmanager/bin"
      chmod 750 ${root_dir}
      tar -zxvf  ${root_dir}/go1.13.linux-amd64.tar.gz -C ${root_dir}/
      nohup ${root_dir}/goc server &
      sleep 5s
      export GOPATH=/home/fleetmanager
      sh ${root_dir}/collect_coverage.sh &
  fi
}

function start_service() {
    ./bin/fleetmanager
}

function set_scc() {
    echo "set scc ..."
    ## SCC_PATH need to be same as cce enviroment variable, default is /home/sccSecret
    new_scc_ks_path=${SCC_PATH:-"/home/sccSecret"}
    mkdir -p "$install_path"/conf/security/ks
    chmod -R u+w  "$install_path"/conf/security/ks
    if [ -f "$new_scc_ks_path/primary.ks" ]; then
        base64 -d "$new_scc_ks_path"/primary.ks > "$install_path"/conf/security/ks/primary.ks 2>/dev/null
        res=$?
        [ $res -ne 0 ] && rm -f $install_path/conf/security/ks/primary.ks && cat "$new_scc_ks_path"/primary.ks > "$install_path"/conf/security/ks/primary.ks
    fi
    if [ -f "$new_scc_ks_path/standby.ks" ]; then
        base64 -d "$new_scc_ks_path"/standby.ks > "$install_path"/conf/security/ks/standby.ks 2>/dev/null
        res=$?
        [ $res -ne 0 ] && rm -f $install_path/conf/security/ks/standby.ks && cat "$new_scc_ks_path"/standby.ks > "$install_path"/conf/security/ks/standby.ks
    fi
    chmod -R u-w  "$install_path"/conf/security/ks
    echo "set scc finish..."
}

function main() {
    set_scc
    start_filebeat
    start_coverage_report
    start_service
}

main
