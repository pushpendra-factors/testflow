FROM python:3.7-alpine

ENV PYTHONPATH "${PYTHONPATH}:/usr/local/var/factors"

RUN mkdir -p /usr/local/var/factors

WORKDIR /usr/local/var/factors
COPY . ./

RUN apk add --no-cache --virtual .build-deps gcc libc-dev libxslt-dev && \
    apk add --no-cache libxslt && \
    pip install --no-cache-dir lxml>=3.5.0 && \
    pip install -r requirements.txt && \
    apk del .build-deps

ENTRYPOINT ["python", "app.py"]
EXPOSE 8091