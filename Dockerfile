FROM alpine AS builder

WORKDIR /boundary

RUN apk update && \
    apk add curl unzip

RUN curl -o boundary-worker.zip https://releases.hashicorp.com/boundary-worker/0.12.3+hcp/boundary-worker_0.12.3+hcp_linux_amd64.zip && \
    unzip boundary-worker.zip && \
    chmod +x boundary-worker


FROM alpine
WORKDIR /boundary
COPY --from=builder /boundary/boundary-worker boundary-worker

RUN apk update && \
        apk add curl jq

EXPOSE 9200

COPY worker.hcl worker.hcl
COPY entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./boundary-worker", "server", "-config=./worker.hcl"]