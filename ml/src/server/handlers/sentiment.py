import logging
from fastapi import APIRouter, HTTPException, Request

from src.server.schemas import (
    SentimentRequest,
    SentimentBatchRequest,
    SentimentResponse,
    SentimentBatchResponse,
    SentimentScores,
)

logger = logging.getLogger(__name__)

router = APIRouter()


def _get_sentiment_model(request: Request):
    model = request.app.state.sentiment_model
    if model is None:
        raise HTTPException(status_code=503, detail="Sentiment model not loaded")
    return model


@router.post("/sentiment", response_model=SentimentResponse)
async def sentiment(req: SentimentRequest, request: Request):
    model = _get_sentiment_model(request)
    try:
        result = model.predict(req.text, req.brand_name)
    except Exception as e:
        logger.exception("Prediction failed")
        raise HTTPException(status_code=503, detail=str(e))

    return SentimentResponse(
        aspect=result["aspect"],
        label=result["label"],
        confidence=result["confidence"],
        scores=SentimentScores(**result["scores"]),
    )


@router.post("/sentiment/batch", response_model=SentimentBatchResponse)
async def sentiment_batch(req: SentimentBatchRequest, request: Request):
    model = _get_sentiment_model(request)
    texts = [item.text for item in req.items]
    brand_names = [item.brand_name for item in req.items]
    try:
        results = model.predict_batch(texts, brand_names)
    except Exception as e:
        logger.exception("Batch prediction failed")
        raise HTTPException(status_code=500, detail=str(e))

    return SentimentBatchResponse(
        results=[
            SentimentResponse(
                aspect=r["aspect"],
                label=r["label"],
                confidence=r["confidence"],
                scores=SentimentScores(**r["scores"]),
            )
            for r in results
        ]
    )
