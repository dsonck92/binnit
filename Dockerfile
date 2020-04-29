FROM golang:1.14.2 as builder

ENV BUILD_DIR /build

ARG UID=1000

RUN mkdir -p ${BUILD_DIR}
WORKDIR ${BUILD_DIR}

COPY go.* ./
RUN go mod download
COPY . .

RUN go test -cover ./...
RUN CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -o /binnit /build/bin/binnit

RUN adduser --system --no-create-home --uid $UID --shell /usr/sbin/nologin static
RUN setcap cap_net_bind_service=+ep /binnit

FROM scratch
COPY --from=builder /binnit /
COPY --from=builder /etc/passwd /etc/passwd

COPY ./conf/binnit.cfg /conf/
COPY ./tpl/* /tpl/
COPY ./static/* /static/

VOLUME /log
VOLUME /paste

# Set our port
EXPOSE 80
WORKDIR /

# Run the binary.
ENTRYPOINT ["/binnit","-c","/conf/binnit.cfg"]
