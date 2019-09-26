ARG IMAGE=centos:centos7
FROM $IMAGE
ARG CERT_PATH

LABEL RUN="docker run -d -p 9000:9000 -v <silo db/log location>:/usr/silo silo" \
    SOURCE="https://github.com/ngageoint/seed-silo" \
    DESCRIPTION="seed-silo api" \
    CLASSIFICATION="UNCLASSIFIED"


WORKDIR /silo
COPY silo /silo

RUN mkdir /usr/silo

# Our app will run on port 9000
EXPOSE 9000

# Start silo
CMD ./silo
