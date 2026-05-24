FROM python:3.12-slim
WORKDIR /app

COPY dashboard/frontend/dist ./dashboard/frontend/dist
COPY dashboard/backend ./dashboard/backend
COPY packages/python/src/dusk ./dusk

RUN pip install --no-cache-dir fastapi uvicorn aiosqlite "redis[asyncio]" pyyaml

EXPOSE 9001

CMD ["uvicorn", "dashboard.backend.server:app", "--host", "0.0.0.0", "--port", "9001"]
