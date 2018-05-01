FROM scratch
MAINTAINER Brian Hechinger <wonko@4amlunch.net>

ADD role-srv-linux-amd64 role-srv
VOLUME /etc/chremoas

ENTRYPOINT ["/role-srv", "--configuration_file", "/etc/chremoas/auth-bot.yaml"]