from __future__ import annotations

import time
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.server.handlers import health_router, sentiment_router, relevance_router, clusterization_router
from src.models.sentimental.model import AspectSentimentModel
from src.models.relevance.model import CompanyRelevanceModel
from src.models.clusterisation.model import TextClusterer

logger = logging.getLogger(__name__)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Loading models...")
    start = time.time()

    try:
        sentiment_model = AspectSentimentModel.load_production()
    except Exception:
        logger.fatal("Production sentiment model not found")
        raise Exception("Production sentiment model not found")
    sentiment_model.eval_mode()

    try:
        relevance_model = CompanyRelevanceModel.load_production()
    except Exception:
        logger.fatal("Production relevance model not found")
        raise Exception("Production relevance model not found")
    relevance_model.eval_mode()

    try:
        clusterization_model = TextClusterer.load_production()
    except Exception:
        logger.fatal("Production clusterization model not found")
        raise Exception("Production clusterization model not found")

    app.state.sentiment_model = sentiment_model
    app.state.relevance_model = relevance_model
    app.state.clusterization_model = clusterization_model

    logger.info(f"Models ready in {time.time() - start:.1f}s")
    yield
    logger.info("Shutting down, releasing models")
    app.state.sentiment_model = None
    app.state.relevance_model = None
    app.state.clusterization_model = None


app = FastAPI(
    title="Aspect Sentiment & Relevance API",
    version="1.0.0",
    description="ABSA — определение тональности по аспектам и релевантности компаний",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(health_router)
app.include_router(sentiment_router)
app.include_router(relevance_router)
app.include_router(clusterization_router)


if __name__ == "__main__":
    import uvicorn
    import yaml
    from pathlib import Path

    config_path = Path("src/server/configs/server_config.yml")
    port = 8000

    if config_path.exists():
        with open(config_path) as f:
            cfg = yaml.safe_load(f)
            port = cfg.get("port", port)

    logger.info(f"Starting server on port {port}")

    uvicorn.run(
        "src.server.server:app",
        host="0.0.0.0",
        port=port,
        log_level="info",
    )
