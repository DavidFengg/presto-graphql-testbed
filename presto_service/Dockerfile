# latest versions of centos do not include python automatically installed
FROM centos:7
RUN yum -y install java-1.8.0-openjdk
RUN mkdir /data
RUN mkdir -p /srv/presto/etc/catalog
WORKDIR /srv/presto

RUN curl -O https://repo1.maven.org/maven2/com/facebook/presto/presto-server/0.229/presto-server-0.229.tar.gz
RUN tar xvfz presto-server-0.229.tar.gz

# add rest-http connector to plugin
RUN mkdir /srv/presto/presto-server-0.229/plugin/rest-http/
COPY rest-plugin/*.jar /srv/presto/presto-server-0.229/plugin/rest-http/

WORKDIR /srv/presto/presto-server-0.229
COPY etc/* /srv/presto/presto-server-0.229/etc/
COPY etc/catalog/* /srv/presto/presto-server-0.229/etc/catalog/
EXPOSE 8080
CMD /srv/presto/presto-server-0.229/bin/launcher run
