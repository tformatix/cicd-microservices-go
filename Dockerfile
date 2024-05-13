FROM golang:1.22.2-alpine

# Set maintainer label: maintainer=[YOUR-EMAIL]
LABEL maintainer="tobias@tformatix.at"

# Set working directory: `/src`
WORKDIR /src

# Copy local files to the working directory
COPY *.go .
COPY *.mod .
COPY *.sum .

# Build the GO app as myapp binary and move it to /usr/

RUN go build -o /usr/myapp ./...

#Expose port 8010
EXPOSE 8010

# Run the service myapp when a container of this image is launched
CMD ["/usr/myapp"]

