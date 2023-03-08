FROM registry-cbu.huawei.com/mpc/mpc-base:2.9.82

RUN useradd -d /home/appgateway -u 1024 -ms /bin/bash appgateway
RUN chown -R appgateway:appgateway /tmp
RUN chmod 750 /tmp

ADD bin  /home/appgateway/bin
ADD conf  /home/appgateway/conf

# 组件日志
RUN mkdir -p /home/appgateway/log/run
# 统计日志，用以数据分析
RUN mkdir -p /home/appgateway/log/monitor
# api监控日志，包括调用者、被调用接口、返回值
RUN mkdir -p /home/appgateway/log/access
# 安全日志
RUN mkdir -p /home/appgateway/log/security
# 调用外部服务的日志
RUN mkdir -p /home/appgateway/log/service_call
# 加密组件日志
RUN mkdir -p /home/appgateway/log/scc
ADD conf/security/logs /home/appgateway/log/scc/logs

RUN chown -R appgateway:appgateway /home/appgateway/*
RUN chown -R appgateway:appgateway /home/base/
RUN chown -R appgateway:appgateway /usr/local/seccomponent

RUN find /home/appgateway/conf/ -type f | xargs chmod ug-wx
RUN find /home/appgateway/bin/ -type f | xargs chmod ug-w
RUN find /home/appgateway/bin/ -type f | xargs chmod ug+x
RUN find /home/appgateway/bin/ -type d | xargs chmod ug-w
RUN chmod 640 /home/appgateway/conf/security/logger.conf
RUN chmod ug=rx /home/appgateway/bin/run.sh
RUN chmod 550 /home/base/filebeat/filebeat
RUN echo "ALWAYS_SET_PATH=yes" > /etc/default/su

EXPOSE 60003
USER appgateway
WORKDIR /home/appgateway

CMD  ["sh","/home/appgateway/bin/run.sh"]
