"""Example: FastAPI app with dusk deprecation middleware."""
from fastapi import FastAPI
from dusk.adapters.fastapi import DuskMiddleware

app = FastAPI()
app.add_middleware(DuskMiddleware, config_path="dusk.yaml")


@app.get("/api/v1/users")
async def list_users_v1():
    return {"users": [], "note": "deprecated — use /api/v2/users"}


@app.get("/api/v2/users")
async def list_users_v2():
    return {"users": [], "page": 1, "total": 0}


@app.get("/api/v1/orders/{order_id}")
async def get_order_v1(order_id: str):
    return {"order_id": order_id}


@app.get("/api/v2/orders/{order_id}")
async def get_order_v2(order_id: str):
    return {"order_id": order_id}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
