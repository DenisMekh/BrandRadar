import logging
from fastapi import APIRouter, HTTPException, Request

from src.server.schemas import (
    RelevanceRequest,
    RelevanceBatchRequest,
    RelevanceResponse,
    RelevanceBatchResponse,
    RelevanceScores,
)

logger = logging.getLogger(__name__)

router = APIRouter()


def _get_relevance_model(request: Request):
    model = request.app.state.relevance_model
    if model is None:
        raise HTTPException(status_code=503, detail="Relevance model not loaded")
    return model


@router.post("/relevance", response_model=RelevanceResponse)
async def relevance(req: RelevanceRequest, request: Request):
    model = _get_relevance_model(request)
    try:
        result = model.predict(req.text, req.brand, keywords=req.keywords)
    except Exception as e:
        logger.exception("Prediction failed")
        raise HTTPException(status_code=503, detail=str(e))

    return RelevanceResponse(
        brand=result["company"],
        label=result["label"],
        is_relevant=result["is_relevant"],
        confidence=result["confidence"],
        scores=RelevanceScores(**result["scores"]),
    )


@router.post("/relevance/batch", response_model=RelevanceBatchResponse)
async def relevance_batch(req: RelevanceBatchRequest, request: Request):
    model = _get_relevance_model(request)
    texts = [item.text for item in req.items]
    company_names = [item.brand for item in req.items]
    keywords_list = [item.keywords for item in req.items]
    try:
        results = model.predict_batch(texts, company_names, keywords=keywords_list)
    except Exception as e:
        logger.exception("Batch prediction failed")
        raise HTTPException(status_code=500, detail=str(e))

    return RelevanceBatchResponse(
        results=[
            RelevanceResponse(
                brand=r["company"],
                label=r["label"],
                is_relevant=r["is_relevant"],
                confidence=r["confidence"],
                scores=RelevanceScores(**r["scores"]),
            )
            for r in results
        ]
    )
