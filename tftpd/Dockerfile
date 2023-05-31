#FROM ubuntu:latest
#RUN useradd -u 65532 scratch

FROM scratch
add tftpd /
#COPY --from=0 /etc/passwd /etc/passwd
#USER scratch

ENTRYPOINT ["/tftpd"]
