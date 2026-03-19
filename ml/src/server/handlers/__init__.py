from src.server.handlers.health import router as health_router
from src.server.handlers.sentiment import router as sentiment_router
from src.server.handlers.relevance import router as relevance_router
from src.server.handlers.clusterization import router as clusterization_router

__all__ = ["health_router", "sentiment_router", "relevance_router", "clusterization_router"]
