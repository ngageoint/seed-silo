FROM docker.platform.cloud.coe.ic.gov/centos:7

LABEL VERSION="0.2.0" \
    RUN="docker run -d -p 9000:9000 -p 80:80 -v <silo db/log location>:/usr/silo docker.platform.cloud.coe.ic.gov/nga-r-dev/silo" \
    SOURCE="https://sc.appdev.proj.coe.ic.gov/dayton-giat/silo" \
    DESCRIPTION="seed-silo api" \
    CLASSIFICATION="UNCLASSIFIED"

RUN mkdir -p silo \
    && update-ca-trust enable \
    && wget http://pki-ldap.ismc.ic.gov/ -r -A *.cer -nd -nv -P /etc/pki/ca-trust/source/anchors/ \
    && update-ca-trust extract

WORKDIR /silo
COPY silo /silo

# Our app will run on port 9000
EXPOSE 9000
# apache will be running on port 80 to handle CORS
EXPOSE 80

# Start silo
CMD silo/silo