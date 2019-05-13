ARG IMAGE=centos:centos7
FROM $IMAGE
ARG CERT_PATH

LABEL RUN="docker run -d -p 9000:9000 -p 80:80 -v <silo db/log location>:/usr/silo silo" \
    SOURCE="https://github.com/ngageoint/seed-silo" \
    DESCRIPTION="seed-silo api" \
    CLASSIFICATION="UNCLASSIFIED"

RUN yum -y install httpd; systemctl enable httpd.service

WORKDIR /silo
COPY silo /silo

COPY httpd-silo.conf /etc/httpd/conf.d/

# Get root certs, if certs arg is present
RUN if [ "x$CERT_PATH" != "x" ] ; then yum install -y wget && update-ca-trust enable && wget $CERT_PATH -r -A *.cer -nd -nv -P /etc/pki/ca-trust/source/anchors/ && update-ca-trust extract; fi

# Our app will run on port 9000
EXPOSE 9000
# apache will be running on port 80 to handle CORS
EXPOSE 80

# Start silo
CMD ./silo
