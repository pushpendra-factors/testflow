FROM python:3.7-alpine

ENV PYTHONPATH "${PYTHONPATH}:/usr/local/var/factors"

RUN mkdir -p /usr/local/var/factors

WORKDIR /usr/local/var/factors
COPY . ./

RUN apk add --no-cache --virtual .build-deps gcc g++ libxslt-dev libffi-dev && \
    apk add --no-cache libxslt libffi && \
    pip install --no-cache-dir lxml>=3.5.0 && \
    pip install -r adwords_requirements.txt

ENTRYPOINT ["python", "app.py"]
EXPOSE 8091