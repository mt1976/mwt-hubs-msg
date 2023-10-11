## We specify the base image we need for our go application
FROM golang:1.19
LABEL version="1.0"
LABEL name="Proteus Hub"
LABEL description="A simple hub for Proteus"
LABEL maintainer="Matt Townsend"
LABEL maintainer_email="mt76@gmx.com"
## We create an /app directory within our
## image that will hold our application source
## files
## We copy everything in the root directory
## into our /app directory
COPY . /app
## We specify that we now wish to execute
## any further commands inside our /app
## directory
WORKDIR /app
## We specify the command that we wish to
## execute inside our image
RUN go mod download
## we run go build to compile the binary
## executable of our Go program
RUN go build -o main .
# Expose port 8111 to the outside world
EXPOSE 8111
## Our start command which kicks off
## our newly created binary executable
CMD ["/app/main"]
