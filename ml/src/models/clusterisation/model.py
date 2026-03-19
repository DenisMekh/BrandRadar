from __future__ import annotations

import sys
import yaml
import json
import logging
import re
import string
import tempfile
from pathlib import Path
from typing import Optional, Dict, List, Union
from dataclasses import dataclass

import numpy as np
import torch
import hdbscan
import umap

from sentence_transformers import SentenceTransformer
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import silhouette_score

sys.path.insert(0, str(Path(__file__).parent.parent.parent))
from src.s3config import S3Config

logger = logging.getLogger(__name__)


def _write_lines(path: Path, lines: List[str]):
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        for x in lines:
            x = (x or "").strip()
            if x:
                f.write(x + "\n")


def _read_lines(path: Path) -> List[str]:
    if not path.exists():
        return []
    out = []
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            s = line.strip()
            if s and not s.startswith("#"):
                out.append(s)
    return out


@dataclass
class ClusterResult:
    cluster_id: int
    messages: List[str]
    indices: List[int]
    topic_keywords: List[str]
    representative_message: str
    size: int
    centroid: Optional[np.ndarray] = None


@dataclass
class ClusteringOutput:
    clusters: List[ClusterResult]
    noise_messages: List[str]
    noise_indices: List[int]
    embeddings: Optional[np.ndarray] = None
    silhouette: Optional[float] = None

    def to_dict(self) -> dict:
        return {
            "num_clusters": len(self.clusters),
            "num_noise": len(self.noise_messages),
            "silhouette_score": self.silhouette,
            "clusters": [
                {
                    "cluster_id": c.cluster_id,
                    "size": c.size,
                    "keywords": c.topic_keywords,
                    "representative": c.representative_message,
                    "messages": c.messages,
                }
                for c in self.clusters
            ],
            "noise": self.noise_messages,
        }


DEFAULT_STOPWORDS = [
    "и", "в", "во", "не", "что", "он", "на", "я", "с", "со", "как", "а", "то",
    "все", "она", "так", "его", "но", "да", "ты", "к", "у", "же", "вы", "за",
    "бы", "по", "только", "ее", "мне", "было", "вот", "от", "меня", "еще",
    "нет", "о", "из", "ему", "теперь", "когда", "даже", "ну", "вдруг", "ли",
    "если", "уже", "или", "ни", "быть", "был", "него", "до", "вас", "нибудь",
    "опять", "уж", "вам", "ведь", "там", "потом", "себя", "ничего", "ей",
    "может", "они", "тут", "где", "есть", "надо", "ней", "для", "мы", "тебя",
    "их", "чем", "была", "сам", "чтоб", "без", "будто", "чего", "раз", "тоже",
    "себе", "под", "будет", "тогда", "кто", "этот", "того", "потому", "этого",
    "какой", "совсем", "ним", "здесь", "этом", "один", "почти", "мой", "тем",
    "чтобы", "нее", "сейчас", "были", "куда", "зачем", "всех", "никогда",
    "можно", "при", "наконец", "два", "об", "другой", "хоть", "после", "над",
    "больше", "тот", "через", "эти", "нас", "про", "всего", "них", "какая",
    "много", "разве", "три", "эту", "моя", "впрочем", "хорошо", "свою", "этой",
    "перед", "иногда", "лучше", "чуть", "том", "нельзя", "такой", "им", "более",
    "всегда", "конечно", "всю", "между",
    "банк", "банка", "банке",
    "тбанк", "т", "т‑банк", "т-банк", "т банк", "tbank", "t-bank",
    "тинькофф", "tinkoff",
    "приложение",
]

DEFAULT_REMOVE_TOKENS = [
    "Т-Банк", "т-банк", "т банк", "t-bank", "t bank", "tbank",
    "Тинькофф", "tinkoff",
]


class TextClusterer:
    """
    SBERT + UMAP + HDBSCAN кластеризация текстов.
    Загрузка: HuggingFace / локальная папка / S3.
    Конфигурация: YAML или параметры __init__.
    """

    def __init__(
        self,
            model_name: str = "ai-forever/sbert_large_nlu_ru",
        max_length: int = 256,
        normalize_embeddings: bool = True,
        batch_size: int = 32,
        device: Optional[str] = None,
        show_progress_bar: bool = False,

            stopwords: Optional[List[str]] = None,
        remove_tokens: Optional[List[str]] = None,
        preprocess_lowercase: bool = True,
        preprocess_remove_punctuation: bool = True,
        preprocess_remove_stopwords: bool = True,
        preprocess_min_length: int = 3,

            use_umap: bool = True,
        umap_n_neighbors: int = 15,
        umap_n_components: int = 5,
        umap_min_dist: float = 0.0,
        umap_metric: str = "cosine",

            tfidf_max_features: int = 3000,
        tfidf_ngram_range: tuple = (1, 2),
        tfidf_min_df: int = 1,
        tfidf_max_df: float = 0.95,
        topic_keywords_top_n: int = 6,
    ):
        self.model_name = model_name
        self.max_length = max_length
        self.normalize_embeddings = normalize_embeddings
        self.batch_size = batch_size
        self.show_progress_bar = show_progress_bar
        self.device = device or ("cuda" if torch.cuda.is_available() else "cpu")

        self.preprocess_lowercase = preprocess_lowercase
        self.preprocess_remove_punctuation = preprocess_remove_punctuation
        self.preprocess_remove_stopwords = preprocess_remove_stopwords
        self.preprocess_min_length = preprocess_min_length
        self.stopwords: set[str] = set(
            s.lower() for s in (stopwords if stopwords is not None else DEFAULT_STOPWORDS)
        )
        self.remove_tokens: List[str] = list(
            remove_tokens if remove_tokens is not None else DEFAULT_REMOVE_TOKENS
        )

        self.use_umap = use_umap
        self.umap_n_neighbors = umap_n_neighbors
        self.umap_n_components = umap_n_components
        self.umap_min_dist = umap_min_dist
        self.umap_metric = umap_metric

        self.tfidf_max_features = tfidf_max_features
        self.tfidf_ngram_range = tfidf_ngram_range
        self.tfidf_min_df = tfidf_min_df
        self.tfidf_max_df = tfidf_max_df
        self.topic_keywords_top_n = topic_keywords_top_n

        logger.info("Initializing TextClusterer on %s", self.device)
        logger.info("Model: %s", model_name)

        self.model = SentenceTransformer(model_name, device=self.device)
        self.model.max_seq_length = max_length

        self.tfidf = TfidfVectorizer(
            max_features=tfidf_max_features,
            stop_words=list(self.stopwords),
            ngram_range=tfidf_ngram_range,
            min_df=tfidf_min_df,
            max_df=tfidf_max_df,
        )

        total_params = sum(p.numel() for p in self.model.parameters())
        logger.info("Model params: %.2fM | Device: %s", total_params / 1e6, self.device)

    @classmethod
    def from_config(cls, config: dict, base_dir: Optional[Union[str, Path]] = None) -> "TextClusterer":
        base_dir = Path(base_dir) if base_dir else Path.cwd()

        m = config.get("model", {})
        p = config.get("preprocess", {})
        u = config.get("umap", {})
        t = config.get("tfidf", {})

        stopwords = None
        sw_path = p.get("stopwords_path")
        sw_inline = p.get("stopwords")
        if sw_path:
            stopwords = _read_lines(base_dir / sw_path)
        elif sw_inline and isinstance(sw_inline, list):
            stopwords = sw_inline

        remove_tokens = None
        rt_path = p.get("remove_tokens_path")
        rt_inline = p.get("remove_tokens")
        if rt_path:
            remove_tokens = _read_lines(base_dir / rt_path)
        elif rt_inline and isinstance(rt_inline, list):
            remove_tokens = rt_inline

        ngram_range = tuple(t.get("ngram_range", [1, 2]))

        return cls(
            model_name=m.get("name", "ai-forever/sbert_large_nlu_ru"),
            max_length=m.get("max_length", 256),
            batch_size=m.get("batch_size", 32),
            normalize_embeddings=m.get("normalize_embeddings", True),
            device=m.get("device"),
            show_progress_bar=m.get("show_progress_bar", False),

            stopwords=stopwords,
            remove_tokens=remove_tokens,
            preprocess_lowercase=p.get("lowercase", True),
            preprocess_remove_punctuation=p.get("remove_punctuation", True),
            preprocess_remove_stopwords=p.get("remove_stopwords", True),
            preprocess_min_length=p.get("min_length", 3),

            use_umap=u.get("use_umap", True),
            umap_n_neighbors=u.get("n_neighbors", 15),
            umap_n_components=u.get("n_components", 5),
            umap_min_dist=u.get("min_dist", 0.0),
            umap_metric=u.get("metric", "cosine"),

            tfidf_max_features=t.get("max_features", 3000),
            tfidf_ngram_range=ngram_range,
            tfidf_min_df=t.get("min_df", 1),
            tfidf_max_df=t.get("max_df", 0.95),
            topic_keywords_top_n=t.get("topic_keywords_top_n", 6),
        )

    @classmethod
    def from_yaml(cls, path: Union[str, Path]) -> "TextClusterer":
        import yaml
        path = Path(path)
        with open(path, "r", encoding="utf-8") as f:
            config = yaml.safe_load(f)
        return cls.from_config(config, base_dir=path.parent)

    def preprocess_texts(self, texts: List[str]) -> List[str]:
        cleaned = []
        for text in texts:
            if self.preprocess_lowercase:
                text = text.lower()

            for token in self.remove_tokens:
                text = re.sub(re.escape(token.lower()), " ", text, flags=re.IGNORECASE)

            if self.preprocess_remove_punctuation:
                text = text.translate(str.maketrans("", "", string.punctuation + '«»""—–…'))

            text = re.sub(r"\s+", " ", text).strip()

            if self.preprocess_remove_stopwords and self.stopwords:
                words = [w for w in text.split() if w not in self.stopwords]
                text = " ".join(words)

            if self.preprocess_min_length > 0:
                words = [w for w in text.split() if len(w) >= self.preprocess_min_length]
                text = " ".join(words)

            text = re.sub(r"\s+", " ", text).strip()
            if not text:
                text = "пустое сообщение"
            cleaned.append(text)
        return cleaned

    def get_embeddings(self, texts: List[str]) -> np.ndarray:
        emb = self.model.encode(
            texts,
            batch_size=self.batch_size,
            show_progress_bar=self.show_progress_bar,
            normalize_embeddings=self.normalize_embeddings,
            convert_to_numpy=True,
        )
        return emb.astype(np.float32)

    def _reduce_umap(self, embeddings: np.ndarray) -> np.ndarray:
        reducer = umap.UMAP(
            n_neighbors=self.umap_n_neighbors,
            n_components=self.umap_n_components,
            min_dist=self.umap_min_dist,
            metric=self.umap_metric,
            random_state=42,
        )
        return reducer.fit_transform(embeddings).astype(np.float32)

    def cluster(
        self,
        texts: List[str],
        min_cluster_size: int = 3,
        min_samples: Optional[int] = None,
        cluster_selection_epsilon: float = 0.0,
        return_embeddings: bool = False,
    ) -> ClusteringOutput:
        if not texts:
            return ClusteringOutput(clusters=[], noise_messages=[], noise_indices=[])

        cleaned = self.preprocess_texts(texts)

        logger.info("Encoding %d texts...", len(texts))
        embeddings = self.get_embeddings(cleaned)

        features = embeddings
        if self.use_umap and len(texts) >= 10:
            logger.info("UMAP reduction...")
            features = self._reduce_umap(embeddings)

        hdb = hdbscan.HDBSCAN(
            min_cluster_size=min_cluster_size,
            min_samples=min_samples if min_samples is not None else max(2, min_cluster_size // 2),
            metric="euclidean",
            cluster_selection_method="eom",
            cluster_selection_epsilon=cluster_selection_epsilon,
        )
        logger.info("Running HDBSCAN...")
        labels = hdb.fit_predict(features)

        tfidf_matrix = self.tfidf.fit_transform(cleaned)
        feature_names = self.tfidf.get_feature_names_out()

        clusters_dict: Dict[int, Dict] = {}
        noise_messages, noise_indices = [], []

        for idx, (orig, label) in enumerate(zip(texts, labels)):
            if label == -1:
                noise_messages.append(orig)
                noise_indices.append(idx)
            else:
                d = clusters_dict.setdefault(
                    label, {"texts": [], "indices": [], "embeddings": [], "tfidf_rows": []}
                )
                d["texts"].append(orig)
                d["indices"].append(idx)
                d["embeddings"].append(embeddings[idx])
                d["tfidf_rows"].append(idx)

        results: List[ClusterResult] = []
        for cid, data in sorted(clusters_dict.items()):
            cl_emb = np.array(data["embeddings"], dtype=np.float32)
            keywords = self._extract_keywords(tfidf_matrix, feature_names, data["tfidf_rows"])
            representative = self._find_representative(cl_emb, data["texts"])
            centroid = cl_emb.mean(axis=0) if self.normalize_embeddings else None

            results.append(ClusterResult(
                cluster_id=cid,
                messages=data["texts"],
                indices=data["indices"],
                topic_keywords=keywords,
                representative_message=representative,
                size=len(data["texts"]),
                centroid=centroid,
            ))

        sil = None
        valid = labels != -1
        if valid.sum() > 2 and len(set(labels[valid])) > 1 and len(texts) <= 10_000:
            sil = float(silhouette_score(features[valid], labels[valid]))

        return ClusteringOutput(
            clusters=results,
            noise_messages=noise_messages,
            noise_indices=noise_indices,
            embeddings=embeddings if return_embeddings else None,
            silhouette=sil,
        )

    def _extract_keywords(
        self, tfidf_matrix, feature_names: np.ndarray, row_indices: List[int],
    ) -> List[str]:
        if len(row_indices) < 2:
            return []
        sub = tfidf_matrix[row_indices]
        mean = np.asarray(sub.mean(axis=0)).ravel()
        top_idx = mean.argsort()[-self.topic_keywords_top_n :][::-1]
        return [feature_names[i] for i in top_idx if mean[i] > 0]

    def _find_representative(self, embeddings: np.ndarray, texts: List[str]) -> str:
        if len(texts) == 1:
            return texts[0]
        centroid = embeddings.mean(axis=0)
        dists = np.linalg.norm(embeddings - centroid, axis=1)
        return texts[int(dists.argmin())]

    def save(self, path: Union[str, Path]):
        path = Path(path)
        path.mkdir(parents=True, exist_ok=True)
        logger.info("Saving clusterer to %s", path)

        self.model.save(str(path / "sentence_transformer"))

        _write_lines(path / "stopwords.txt", sorted(self.stopwords))
        _write_lines(path / "remove_tokens.txt", self.remove_tokens)

        meta = {
            "model_type": "text_clusterer",
            "model_name": self.model_name,
            "max_length": self.max_length,
            "normalize_embeddings": self.normalize_embeddings,
            "batch_size": self.batch_size,
            "show_progress_bar": self.show_progress_bar,
            "preprocess": {
                "lowercase": self.preprocess_lowercase,
                "remove_punctuation": self.preprocess_remove_punctuation,
                "remove_stopwords": self.preprocess_remove_stopwords,
                "min_length": self.preprocess_min_length,
                "stopwords_file": "stopwords.txt",
                "remove_tokens_file": "remove_tokens.txt",
            },
            "umap": {
                "use_umap": self.use_umap,
                "n_neighbors": self.umap_n_neighbors,
                "n_components": self.umap_n_components,
                "min_dist": self.umap_min_dist,
                "metric": self.umap_metric,
            },
            "tfidf": {
                "max_features": self.tfidf_max_features,
                "ngram_range": list(self.tfidf_ngram_range),
                "min_df": self.tfidf_min_df,
                "max_df": self.tfidf_max_df,
                "topic_keywords_top_n": self.topic_keywords_top_n,
            },
        }

        with open(path / "meta.json", "w", encoding="utf-8") as f:
            json.dump(meta, f, ensure_ascii=False, indent=2)

        logger.info("Clusterer saved to %s", path)

    def load(self, path: Union[str, Path]):
        path = Path(path)
        logger.info("Loading clusterer from %s", path)
        meta = self._load_meta(path)

        self.model_name = meta.get("model_name", self.model_name)
        self.max_length = meta.get("max_length", self.max_length)
        self.normalize_embeddings = meta.get("normalize_embeddings", self.normalize_embeddings)
        self.batch_size = meta.get("batch_size", self.batch_size)
        self.show_progress_bar = meta.get("show_progress_bar", False)

        p = meta.get("preprocess", {})
        self.preprocess_lowercase = p.get("lowercase", True)
        self.preprocess_remove_punctuation = p.get("remove_punctuation", True)
        self.preprocess_remove_stopwords = p.get("remove_stopwords", True)
        self.preprocess_min_length = p.get("min_length", 3)
        self.stopwords = set(
            s.lower() for s in _read_lines(path / p.get("stopwords_file", "stopwords.txt"))
        )
        self.remove_tokens = _read_lines(path / p.get("remove_tokens_file", "remove_tokens.txt"))

        u = meta.get("umap", {})
        self.use_umap = u.get("use_umap", True)
        self.umap_n_neighbors = u.get("n_neighbors", 15)
        self.umap_n_components = u.get("n_components", 5)
        self.umap_min_dist = u.get("min_dist", 0.0)
        self.umap_metric = u.get("metric", "cosine")

        t = meta.get("tfidf", {})
        self.tfidf_max_features = t.get("max_features", 3000)
        self.tfidf_ngram_range = tuple(t.get("ngram_range", [1, 2]))
        self.tfidf_min_df = t.get("min_df", 1)
        self.tfidf_max_df = t.get("max_df", 0.95)
        self.topic_keywords_top_n = t.get("topic_keywords_top_n", 6)

        self.model = SentenceTransformer(str(path / "sentence_transformer"), device=self.device)
        self.model.max_seq_length = self.max_length

        self.tfidf = TfidfVectorizer(
            max_features=self.tfidf_max_features,
            stop_words=list(self.stopwords),
            ngram_range=self.tfidf_ngram_range,
            min_df=self.tfidf_min_df,
            max_df=self.tfidf_max_df,
        )

        logger.info("Clusterer loaded from %s", path)

    @classmethod
    def from_pretrained(
        cls,
        path: str,
        endpoint_url: str = None,
        aws_access_key_id: str = None,
        aws_secret_access_key: str = None,
        region_name: str = None,
    ) -> "TextClusterer":
        if path.startswith("s3://"):
            logger.info("Detected S3 path: %s", path)
            local_path = cls._download_from_s3(
                s3_path=path,
                endpoint_url=endpoint_url,
                aws_access_key_id=aws_access_key_id,
                aws_secret_access_key=aws_secret_access_key,
                region_name=region_name,
            )
        else:
            local_path = Path(path)
            if local_path.is_symlink():
                local_path = local_path.resolve()
            logger.info("Using local path: %s", local_path)

        if not local_path.exists():
            raise FileNotFoundError(f"Model path not found: {local_path}")

        meta = cls._load_meta_static(local_path)
        p = meta.get("preprocess", {})
        u = meta.get("umap", {})
        t = meta.get("tfidf", {})

        stopwords = _read_lines(local_path / p.get("stopwords_file", "stopwords.txt"))
        remove_tokens = _read_lines(local_path / p.get("remove_tokens_file", "remove_tokens.txt"))

        instance = cls(
            model_name=str(local_path / "sentence_transformer"),
            max_length=meta.get("max_length", 256),
            normalize_embeddings=meta.get("normalize_embeddings", True),
            batch_size=meta.get("batch_size", 32),
            device=None,
            show_progress_bar=meta.get("show_progress_bar", False),

            stopwords=stopwords,
            remove_tokens=remove_tokens,
            preprocess_lowercase=p.get("lowercase", True),
            preprocess_remove_punctuation=p.get("remove_punctuation", True),
            preprocess_remove_stopwords=p.get("remove_stopwords", True),
            preprocess_min_length=p.get("min_length", 3),

            use_umap=u.get("use_umap", True),
            umap_n_neighbors=u.get("n_neighbors", 15),
            umap_n_components=u.get("n_components", 5),
            umap_min_dist=u.get("min_dist", 0.0),
            umap_metric=u.get("metric", "cosine"),

            tfidf_max_features=t.get("max_features", 3000),
            tfidf_ngram_range=tuple(t.get("ngram_range", [1, 2])),
            tfidf_min_df=t.get("min_df", 1),
            tfidf_max_df=t.get("max_df", 0.95),
            topic_keywords_top_n=t.get("topic_keywords_top_n", 6),
        )

        instance.model_name = meta.get("model_name", instance.model_name)

        total_params = sum(p.numel() for p in instance.model.parameters())
        logger.info(
            "Clusterer loaded from %s (%.1fM params, device=%s)",
            local_path, total_params / 1e6, instance.device,
        )
        return instance

    @staticmethod
    def _load_meta_static(path: Path) -> dict:
        meta_json = path / "meta.json"
        if not meta_json.exists():
            raise FileNotFoundError(f"meta.json not found in {path}")
        with open(meta_json, encoding="utf-8") as f:
            return json.load(f)

    def _load_meta(self, path: Path) -> dict:
        return self._load_meta_static(path)

    @staticmethod
    def _download_from_s3(
        s3_path: str,
        endpoint_url: str = None,
        aws_access_key_id: str = None,
        aws_secret_access_key: str = None,
        region_name: str = None,
    ) -> Path:
        try:
            import boto3
        except ImportError:
            raise ImportError("pip install boto3")
        try:
            from tqdm import tqdm
        except ImportError:
            tqdm = None

        parts = s3_path.replace("s3://", "").split("/", 1)
        bucket = parts[0]
        prefix = parts[1] if len(parts) > 1 else ""

        session_kwargs = {}
        if aws_access_key_id:
            session_kwargs["aws_access_key_id"] = aws_access_key_id
        if aws_secret_access_key:
            session_kwargs["aws_secret_access_key"] = aws_secret_access_key
        if region_name:
            session_kwargs["region_name"] = region_name

        session = boto3.session.Session(**session_kwargs)
        client_kwargs = {"endpoint_url": endpoint_url} if endpoint_url else {}
        s3 = session.client("s3", **client_kwargs)

        all_objects = []
        continuation_token = None
        while True:
            list_kwargs = {"Bucket": bucket, "Prefix": prefix}
            if continuation_token:
                list_kwargs["ContinuationToken"] = continuation_token
            resp = s3.list_objects_v2(**list_kwargs)
            if "Contents" not in resp:
                break
            all_objects.extend(resp["Contents"])
            if resp.get("IsTruncated"):
                continuation_token = resp["NextContinuationToken"]
            else:
                break

        if not all_objects:
            raise FileNotFoundError(f"No objects at s3://{bucket}/{prefix}")

        files = []
        total_bytes = 0
        for obj in all_objects:
            key = obj["Key"]
            rel = key[len(prefix):].lstrip("/")
            if not rel or rel.endswith("/"):
                continue
            files.append((key, rel, obj["Size"]))
            total_bytes += obj["Size"]

        local_dir = Path(tempfile.mkdtemp(prefix="cluster_model_"))
        progress = tqdm(total=total_bytes, unit="B", unit_scale=True, unit_divisor=1024) if tqdm else None

        for key, rel, _ in files:
            target = local_dir / rel
            target.parent.mkdir(parents=True, exist_ok=True)
            if progress is not None:
                def _make_cb(pbar):
                    def _cb(bytes_transferred):
                        pbar.update(bytes_transferred)
                    return _cb
                s3.download_file(bucket, key, str(target), Callback=_make_cb(progress))
            else:
                s3.download_file(bucket, key, str(target))

        if progress is not None:
            progress.close()

        logger.info("Model downloaded to %s", local_dir)
        return local_dir

    @staticmethod
    def load_production() -> "TextClusterer":
        cfg_path = Path("src/models/clusterisation/configs/prod_config.yml")
        with open(cfg_path, "r", encoding="utf-8") as f:
            model_config = yaml.safe_load(f)
        model_path = model_config["model"]["name"]

        cfg = S3Config()

        return TextClusterer.from_pretrained(
            path=model_path,
            endpoint_url=cfg.S3_ENDPOINT_URL,
            aws_access_key_id=cfg.S3_ACCESS_KEY,
            aws_secret_access_key=cfg.S3_SECRET_KEY,
            region_name=cfg.S3_REGION_NAME
        )

    def eval_mode(self):
        self.model.eval()
