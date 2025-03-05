FROM quay.io/openshift/origin-cli:latest

# copy all collection scripts to /usr/bin
COPY collection-scripts /usr/bin/

RUN mkdir /licenses
COPY ../LICENSE /licenses

ENTRYPOINT /usr/bin/gather
