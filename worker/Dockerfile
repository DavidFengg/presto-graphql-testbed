# latest versions of centos do not include python automatically installed
FROM centos:7
RUN yum -y install java-1.8.0-openjdk
RUN mkdir /data
RUN mkdir -p /srv/presto/etc/catalog
WORKDIR /srv/presto
RUN curl -O https://repo1.maven.org/maven2/com/facebook/presto/presto-server/0.218/presto-server-0.218.tar.gz
RUN tar xvfz presto-server-0.218.tar.gz
WORKDIR /srv/presto/presto-server-0.218
COPY etc/* /srv/presto/presto-server-0.218/etc/
COPY etc/catalog/* /srv/presto/presto-server-0.218/etc/catalog/
EXPOSE 8081
CMD /srv/presto/presto-server-0.218/bin/launcher run
