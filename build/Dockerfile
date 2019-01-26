FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/backup-restore-operator /usr/local/bin/backup-restore-operator
