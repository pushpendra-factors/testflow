FROM python:3.7-alpine

WORKDIR /adwords
COPY * ./
RUN apk add --no-cache --virtual .build-deps gcc libc-dev libxslt-dev && \
    apk add --no-cache libxslt && \
    pip install --no-cache-dir lxml>=3.5.0 && \
    apk del .build-deps && \
    pip install -r requirements.txt

ENTRYPOINT ["python", "app.py"]
EXPOSE 8091