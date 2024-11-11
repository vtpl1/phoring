FROM ubuntu:20.04
WORKDIR /workfiles
COPY ./backend/backend ./backend
ENTRYPOINT ["/workfiles/backend"]
