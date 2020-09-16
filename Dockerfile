FROM golang:1.15.2-alpine3.12 as build_stage
ENV CGO_ENABLED 0
ARG VCS_REF
ARG PACKAGE_NAME

# Create a location in the image for the source code.
RUN mkdir -p /app
WORKDIR /app

# Copy the module files first and then download the dependencies.
COPY go.* ./
#RUN go mod download

# Copy the source code into the image.
COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY tls tls
COPY ui ui
COPY vendor vendor

# Build the admin tool so we can have it in the image.
WORKDIR /app/cmd/${PACKAGE_NAME}-admin
RUN go build

# Build the service binary.
WORKDIR /app/cmd/${PACKAGE_NAME}
RUN go build -ldflags "-X main.build=${VCS_REF}"
# The linker sets 'var build' in main.go to the specified git revision
# See https://golang.org/cmd/link/ for supported linker flags


# Build production image with Go binaries based on Alpine.
FROM alpine:3.12
ARG BUILD_DATE
ARG VCS_REF
ARG PACKAGE_NAME
RUN addgroup -g 1000 -S app && adduser -u 1000 -S app -G app --no-create-home --disabled-password
USER app
COPY --from=build_stage --chown=app:app /app/cmd/${PACKAGE_NAME}-admin/${PACKAGE_NAME}-admin /app/admin
COPY --from=build_stage --chown=app:app /app/cmd/${PACKAGE_NAME}/${PACKAGE_NAME} /app/main
COPY --from=build_stage --chown=app:app /app/ui /app/ui
COPY --from=build_stage --chown=app:app /app/tls /app/tls
WORKDIR /app
CMD ["/app/main"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="${PACKAGE_NAME}" \
      org.opencontainers.image.authors="Andreas <tullo@pm.me>" \
      org.opencontainers.image.source="https://github.com/tullo/snptx/cmd/${PACKAGE_NAME}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.vendor="Amstutz-IT"
