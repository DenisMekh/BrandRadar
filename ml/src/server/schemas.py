from pydantic import BaseModel, Field


class SentimentRequest(BaseModel):
    text: str = Field(..., min_length=1, description="Текст отзыва")
    brand_name: str = Field(..., min_length=1, description="Название бренда")

class SentimentBatchRequest(BaseModel):
    items: list[SentimentRequest] = Field(..., min_length=1, max_length=128)


class SentimentScores(BaseModel):
    positive: float
    negative: float
    neutral: float


class SentimentResponse(BaseModel):
    aspect: str
    label: str
    confidence: float
    scores: SentimentScores


class SentimentBatchResponse(BaseModel):
    results: list[SentimentResponse]


class HealthResponse(BaseModel):
    status: str
    model_loaded: bool
    device: str


class RelevanceRequest(BaseModel):
    text: str = Field(..., min_length=1, description="Текст сообщения")
    brand: str = Field(..., min_length=1, description="Название компании")
    keywords: str | None = Field(default=None, description="Ключевые слова компании (опционально)")


class RelevanceBatchRequest(BaseModel):
    items: list[RelevanceRequest] = Field(..., min_length=1, max_length=128)


class RelevanceScores(BaseModel):
    relevant: float
    irrelevant: float


class RelevanceResponse(BaseModel):
    brand: str
    label: str
    is_relevant: bool
    confidence: float
    scores: RelevanceScores


class RelevanceBatchResponse(BaseModel):
    results: list[RelevanceResponse]


class ClusterizationRequest(BaseModel):
    texts: list[str] = Field(..., min_length=1, max_length=10000, description="Список текстов для кластеризации")
    min_cluster_size: int = Field(default=3, ge=2, description="Минимальный размер кластера")


class ClusterInfo(BaseModel):
    cluster_id: int
    size: int
    keywords: list[str]
    representative: str
    messages: list[str]


class ClusterizationResponse(BaseModel):
    num_clusters: int
    num_noise: int
    silhouette_score: float | None
    clusters: list[ClusterInfo]
    noise: list[str]
