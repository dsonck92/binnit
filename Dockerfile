FROM busybox AS build-env
RUN mkdir /stuff
RUN mkdir /stuff/binnit
RUN mkdir /stuff/binnit/conf
RUN mkdir /stuff/binnit/log
RUN mkdir /stuff/binnit/static
RUN mkdir /stuff/binnit/tpl
RUN mkdir /stuff/binnit/paste

FROM scratch
COPY --from=build-env /stuff /

COPY binnit /binnit/
COPY ./conf/binnit.cfg /binnit/conf/
COPY ./tpl/* /binnit/tpl/
COPY ./static/* /binnit/static/

# Export our volumes so directories
# can be bind mounted
VOLUME /binnit/conf
VOLUME /binnit/log
VOLUME /binnit/static
VOLUME /binnit/tpl
VOLUME /binnit/paste

# Set our port
EXPOSE 80
WORKDIR /binnit

# Run the binary.
ENTRYPOINT ["./binnit","-c","conf/binnit.cfg"]
