FROM golang:1.15.6 as build_service-api
ENV CGO_ENABLED 0
ARG VCS_REF

RUN mkdir -p /service

WORKDIR /service
COPY . .

WORKDIR /service/cmd/service-admin
RUN go build -ldflags "-X main.build=${VCS_REF}"

WORKDIR /service/cmd/service-api
RUN go build -ldflags "-X main.build=${VCS_REF}"


FROM alpine:3.12
ARG BUILD_DATE
ARG VCS_REF
COPY --from=build_service-api /service/private.pem /service/private.pem
COPY --from=build_service-api /service/cmd/service-admin/service-admin /service/admin
COPY --from=build_service-api /service/cmd/service-api/service-api /service/service-api
WORKDIR /service

CMD ["./service-api"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="service-api" \
      org.opencontainers.image.source="https://github.com/santiagoh1997/service-template/cmd/service-api" \
      org.opencontainers.image.revision="${VCS_REF}" \