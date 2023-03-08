FROM registry-cbu.huawei.com/mpc/mpc-base:2.9.82

RUN useradd -d /home/fleetmanager -u 1024 -ms /bin/bash fleetmanager
RUN chown -R fleetmanager:fleetmanager /tmp
RUN chmod 750 /tmp

ADD bin  /home/fleetmanager/bin
ADD conf  /home/fleetmanager/conf

RUN mkdir -p /home/fleetmanager/log/run
RUN mkdir -p /home/fleetmanager/log/monitor
RUN mkdir -p /home/fleetmanager/log/access
RUN mkdir -p /home/fleetmanager/log/service_call
RUN mkdir -p /home/fleetmanager/log/security
RUN mkdir -p /home/fleetmanager/log/scc
ADD ./conf/security/logs /home/fleetmanager/log/scc/logs/

RUN rm -rf /home/fleetmanager/conf/security/logs/*

RUN chown -R fleetmanager:fleetmanager /home/fleetmanager/*
RUN chown -R fleetmanager:fleetmanager /home/base/
RUN chown -R fleetmanager:fleetmanager /usr/local/seccomponent

RUN find /home/fleetmanager/conf/ -type f | xargs chmod ug-wx
RUN find /home/fleetmanager/conf/ -type d | xargs chmod ug-w
RUN find /home/fleetmanager/bin/ -type f | xargs chmod ug-w
RUN find /home/fleetmanager/bin/ -type f | xargs chmod ug+x
RUN find /home/fleetmanager/bin/ -type d | xargs chmod ug-w
RUN find /home/fleetmanager/ -type f | xargs chmod o-rwx
RUN find /home/fleetmanager/ -type d | xargs chmod o-rwx

RUN chmod 550 /home/base/filebeat/filebeat

RUN echo "ALWAYS_SET_PATH=yes" > /etc/default/su

EXPOSE 8080
USER fleetmanager
WORKDIR /home/fleetmanager

CMD  ["sh","/home/fleetmanager/bin/run.sh"]
