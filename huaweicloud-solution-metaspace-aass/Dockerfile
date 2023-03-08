FROM registry-cbu.huawei.com/mpc/mpc-base:2.9.82

RUN useradd -d /home/scase -u 1024 -ms /bin/bash scase
RUN chown -R scase:scase /tmp
RUN chmod 750 /tmp

ADD bin /home/scase/bin
ADD conf  /home/scase/conf

# 组件日志
RUN mkdir -p /home/scase/log/run
# 运维日志，用以数据分析
RUN mkdir -p /home/scase/log/monitor
# api被调用日志，包括调用者、被调用接口、返回值
RUN mkdir -p /home/scase/log/access
# 调用外部服务的日志
RUN mkdir -p /home/scase/log/service-call
# 加密组件日志
RUN mkdir -p /home/scase/log/scc
ADD ./conf/security/logs /home/scase/log/scc/logs/

RUN chown -R scase:scase /home/scase/*
RUN chown -R scase:scase /home/base/
RUN chown -R scase:scase /usr/local/seccomponent

RUN find /home/scase/conf/ -type f | xargs chmod ug-wx
RUN find /home/scase/bin/ -type f | xargs chmod ug-w
RUN find /home/scase/bin/ -type f | xargs chmod ug+x
RUN find /home/scase/bin/ -type d | xargs chmod ug-w
RUN chmod 640 /home/scase/conf/security/logger.conf
RUN chmod ug=rx /home/scase/bin/run.sh
RUN chmod 550 /home/base/filebeat/filebeat
RUN echo "ALWAYS_SET_PATH=yes" > /etc/default/su

EXPOSE 9091
USER scase
WORKDIR /home/scase

CMD ["sh","/home/scase/bin/run.sh"]
