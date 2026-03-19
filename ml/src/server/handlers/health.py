from fastapi import APIRouter, Request

from src.server.schemas import HealthResponse

router = APIRouter()


@router.get("/health", response_model=HealthResponse)
async def health(request: Request):
    sentiment_model = request.app.state.sentiment_model
    return HealthResponse(
        status="ok" if sentiment_model else "unavailable",
        model_loaded=sentiment_model is not None,
        device=str(sentiment_model.device) if sentiment_model else "none",
    )
