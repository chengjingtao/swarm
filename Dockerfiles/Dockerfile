FROM ubuntu:14.04
COPY ./swarm /swarm
#https相关的会用到
COPY ./certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ./tmp /tmp

ENV SWARM_HOST :2375
EXPOSE 2375

VOLUME /.swarm
VOLUME /config.json
ENV TZ Asia/Shanghai
ENTRYPOINT ["/swarm"]
CMD ["--help"]

#sudo docker run -d -v /home/cjt/csdn/container-app/swarm/config.json:/config.json chengjt/swarm --debug manage -H 0.0.0.0:4375 --strategy forever-spread  nodes://192.168.5.55:2375
