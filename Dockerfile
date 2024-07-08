FROM golang:1.21.5

ARG APP_DIR=/app

ENV TODO_PORT=7540
ENV TODO_DBFILE=${APP_DIR}/db/scheduler.db
ENV TODO_PASSWORD=12345
ENV TODO_WEBDIR=${APP_DIR}/web

WORKDIR ${APP_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd
COPY internal ./internal
COPY web ./web

RUN go build -o ${APP_DIR}/go_final_project ./cmd/app/main.go

EXPOSE ${TODO_PORT}

CMD ["./go_final_project"]