FROM alpine AS builder

WORKDIR /boundary

RUN apk update && \
    apk add curl unzip

RUN curl -o boundary.zip https://releases.hashicorp.com/boundary/0.13.1+ent/boundary_0.13.1+ent_linux_amd64.zip && \
    unzip boundary.zip && \
    chmod +x boundary


FROM alpine
WORKDIR /boundary
COPY --from=builder /boundary/boundary boundary

RUN apk update && \
        apk add curl jq

EXPOSE 9200

COPY worker.hcl worker.hcl
COPY entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
CMD ["./boundary", "server", "-config=./worker.hcl"]