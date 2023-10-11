FROM golang:1.21

ADD marmiton/go.mod /src/marmiton/go.mod
ADD marmiton/go.sum /src/marmiton/go.sum

ADD marmiton/db/go.mod /src/marmiton/db/go.mod
ADD marmiton/db/go.sum /src/marmiton/db/go.sum

WORKDIR /src/marmiton
RUN go mod download && cd db && go mod download

ADD marmiton /src/marmiton
RUN go build .

CMD ["./marmiton"]