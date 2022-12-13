FROM golang:alpine as builder

RUN mkdir /build

# add everything from the root directory into the build folder that we just created
ADD . /build/

WORKDIR /build

# compile the entire build folder
RUN go build -o main .

# stage 2
FROM alpine

# create a new user
RUN adduser -D -S -H -h /app appuser

# switch to the newly created user 
USER appuser

# copy everything from our current app into the final product
COPY . /app

# copy all the contents from the build/main folder into the app folder
COPY --from=builder /build/main /app/

WORKDIR /app

# open port 3000 for the application
EXPOSE 3000

CMD ["./main"]
