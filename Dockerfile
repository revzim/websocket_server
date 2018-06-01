# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /Documents/go/src/github.com/imdatdude/ws_play_chat_integration

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get -u github.com/gorilla/websocket
RUN go get -u github.com/gorilla/mux
RUN go run /Documents/go/src/github.com/imdatdude/ws_play_chat_integration/*.go

# Run the outyet command by default when the container starts.
ENTRYPOINT go run /Documents/go/src/github.com/imdatdude/ws_play_chat_integration/*.go

# Document that the service listens on port 8080.
EXPOSE 8080