FROM alpine:3.13.2

WORKDIR /app

ENTRYPOINT ["/app/twbridge"]

COPY twbridge /app/twbridge

RUN chmod 755 /app/twbridge

ENV GOTRACEBACK=crash
