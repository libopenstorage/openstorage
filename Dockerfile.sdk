FROM alpine
MAINTAINER luis@portworx.com

RUN apk update && \
  apk add --no-cache ca-certificates

EXPOSE 9100 9110
ADD ./etc/config/config-fake.yaml /config-fake.yaml
ADD ./_tmp/osd /
ENTRYPOINT ["/osd"]
CMD ["-d", "-f", "/config-fake.yaml"]
