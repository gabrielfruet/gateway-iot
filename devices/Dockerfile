FROM python:3.11-slim

WORKDIR /app

COPY devices /app
COPY requirements.txt /app/requirements.txt

RUN pip install --no-cache-dir -r /app/requirements.txt

