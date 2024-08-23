FROM golang:1.23.0 AS build

WORKDIR /app

COPY app .

RUN pwd && cat go.mod && ls -lathr && go mod download
RUN CGO_ENABLED=0 go build -o quizengine .

FROM ubuntu
WORKDIR /app
COPY --from=build /app/quizengine /app/quizengine
COPY --from=build /app/templates /app/templates
COPY --from=build /app/static /app/static
COPY --from=build /app/config.yaml /app/config.yaml

EXPOSE 8080
CMD ["/app/quizengine"]
