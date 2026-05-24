FROM python:3.12-slim
WORKDIR /app

COPY dashboard/dist ./dashboard/dist
COPY dashboard_server ./dashboard_server
COPY core ./core

RUN pip install --no-cache-dir fastapi uvicorn aiosqlite "redis[asyncio]" pyyaml

EXPOSE 9001

CMD ["uvicorn", "dashboard_server.server:app", "--host", "0.0.0.0", "--port", "9001"]
