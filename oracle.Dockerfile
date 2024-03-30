FROM alpine
COPY oracle /oracle
ENTRYPOINT ["/oracle"]