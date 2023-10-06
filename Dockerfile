FROM golang:1.21

ADD marmiton /src/marmiton

WORKDIR /src/marmiton

RUN go get .
RUN go build .

CMD ["./marmiton"]