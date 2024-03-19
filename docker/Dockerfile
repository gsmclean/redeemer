#parent Image
FROM golang:1.22.0-bullseye

WORKDIR /app

COPY ./main .

RUN go mod download

RUN go build -o /goRedeem

EXPOSE 8083

CMD [ "/goRedeem"]
