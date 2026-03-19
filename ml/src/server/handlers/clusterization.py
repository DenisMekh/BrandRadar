import logging
from fastapi import APIRouter, HTTPException, Request

from src.models.clusterisation.model import TextClusterer
from src.server.schemas import (
    ClusterizationRequest,
    ClusterizationResponse,
    ClusterInfo,
)

logger = logging.getLogger(__name__)

router = APIRouter()


def _get_clusterization_model(request: Request) -> TextClusterer:
    model = request.app.state.clusterization_model
    if model is None:
        raise HTTPException(status_code=503, detail="Clusterization model not loaded")
    return model


@router.post("/clusterization", response_model=ClusterizationResponse)
async def clusterization(req: ClusterizationRequest, request: Request):
    model = _get_clusterization_model(request)
    try:
        result = model.cluster(
            texts=req.texts,
            min_cluster_size=req.min_cluster_size,
        )
    except Exception as e:
        logger.exception("Clusterization failed")
        raise HTTPException(status_code=500, detail=str(e))

    return ClusterizationResponse(
        num_clusters=len(result.clusters),
        num_noise=len(result.noise_messages),
        silhouette_score=result.silhouette,
        clusters=[
            ClusterInfo(
                cluster_id=c.cluster_id,
                size=c.size,
                keywords=c.topic_keywords,
                representative=c.representative_message,
                messages=c.messages,
            )
            for c in result.clusters
        ],
        noise=result.noise_messages,
    )
