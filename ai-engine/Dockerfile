FROM python:3.12-slim

WORKDIR /app

COPY ai-engine/pyproject.toml .
COPY ai-engine/src/ src/

RUN pip install --no-cache-dir .

EXPOSE 8086
CMD ["uvicorn", "qsgw_ai.api.app:app", "--host", "0.0.0.0", "--port", "8086"]
