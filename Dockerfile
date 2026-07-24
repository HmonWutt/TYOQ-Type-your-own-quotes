FROM python:3.12-slim
WORKDIR /app
COPY app/main.py .
COPY app/db.py .
CMD ["python", "-u", "main.py"]
