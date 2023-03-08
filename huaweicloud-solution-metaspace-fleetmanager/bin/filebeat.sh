#!/bin/bash

function start_client() {
  echo "filebeat config file $1, path.data=/home/base/filebeat/data$2"

  /home/base/filebeat/filebeat --path.data /home/base/filebeat/data$2 -e -c $1 >/dev/null 2>&1 &
  if [ $? == 0 ]; then
      echo "start filebeat ok, file=$1."
  else
      echo "start filebeat failed, file=$1."
  fi
}

function start_filebeat_clients() {
  if [ "${Use_Filebeat}" == "true" ];then
    echo "start filebeat..."
    local config_file
    local idx=1
    if [ -n "$FILEBEAT_CONF" ]; then
        for config_file in $(echo $FILEBEAT_CONF | sed "s/:/ /g")
        do
            start_client $config_file $idx
            idx=$((idx+1))
        done
    else
        config_file=/home/fleetmanager/conf/filebeat.yml
        sed -i "s/EVN_REGION/$REGION/g" $config_file
        start_client $config_file $idx
    fi
  fi
}

start_filebeat_clients