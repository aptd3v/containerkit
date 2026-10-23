# syntax=docker/dockerfile:1
# stage a static page, then serve it with busybox httpd
FROM alpine:latest AS build
WORKDIR /out
RUN printf 'Hello from containerkit\n' > index.html && \
	chmod 644 index.html
FROM alpine:latest
RUN apk add --no-cache busybox-extras
COPY --from=build /out/index.html /www/index.html
LABEL example=dockerfile
EXPOSE 8080
USER nobody
HEALTHCHECK --interval=5s --timeout=2s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8080/ >/dev/null
ENTRYPOINT ["/usr/sbin/httpd", "-f", "-p", "8080", "-h", "/www"]
