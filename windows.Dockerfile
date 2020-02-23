FROM mediadevices-base

ENV TARGET_OS=windows
ENV PYTHONUNBUFFERED=1

COPY scripts /usr/bin/ 

RUN install-dependencies