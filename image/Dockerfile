FROM trinodb/trino:356

USER root

ADD  https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 /usr/bin/jq

COPY ./jars/jmx_prometheus_javaagent-0.13.0.jar /usr/lib/trino/lib

RUN rm /usr/lib/trino/lib/trino-main-356.jar

COPY ./jars/trino-main-356.jar /usr/lib/trino/lib/trino-main-356.jar

RUN chmod 755 /usr/bin/jq

USER trino

ENV TRINO_HOME=/usr/lib/trino

