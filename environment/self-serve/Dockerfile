FROM golang:1.15 as build
WORKDIR /build
COPY . /build
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static
COPY --from=build /build/self-serve /self-serve
COPY mvp.css /mvp.css
COPY favicon.ico /favicon.ico
CMD ["/self-serve"]