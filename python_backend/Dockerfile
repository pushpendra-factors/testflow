FROM python:3.8-slim

ENV PYTHONPATH "${PYTHONPATH}:/usr/local/var/factors:/usr/local/var/factors/chat_factors:/usr/local/var/factors/chat_factors/chatgpt_poc:/usr/local/var/factors/chat_factors/chatgpt_poc/chat_jobs/"

RUN mkdir -p /usr/local/var/factors

WORKDIR /usr/local/var/factors
COPY . ./

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    libopenblas-dev \
    liblapack-dev \
    libblas-dev \
    libatlas-base-dev \
    libffi-dev \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*


RUN  pip install -r chat_requirements.txt
RUN pip install -r adwords_requirements.txt
ENTRYPOINT ["python", "app.py"]
EXPOSE 8091