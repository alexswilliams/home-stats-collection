FROM python:3.7-slim

WORKDIR /usr/src/app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt && apt-get update && apt-get install curl -y && rm -rf /var/lib/apt/lists/*

COPY main.py entry.sh ./

CMD [ "bash", "./entry.sh" ]