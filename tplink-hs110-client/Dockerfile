FROM golang:1.12-alpine as build
WORKDIR /go/src/app
COPY src ./
RUN CGO_ENABLED=0 go build -o scraper -ldflags="-s -w" -tags netgo -a ./main

FROM scratch
WORKDIR /app
COPY --from=build /go/src/app/scraper ./

CMD [ "./scraper" ]
