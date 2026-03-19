from __future__ import annotations

import sys
import json
import logging
import torch
import torch.nn as nn
from transformers import AutoTokenizer, AutoModel
from pathlib import Path
from tqdm import tqdm
import boto3
import tempfile
import yaml

sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent))
from src.s3config import S3Config


logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)


class AspectSentimentModel:
    LABEL2ID = {"positive": 0, "negative": 1, "neutral": 2}
    ID2LABEL = {v: k for k, v in LABEL2ID.items()}

    def __init__(
        self,
        model_name: str = "cointegrated/rubert-tiny2",
        num_labels: int = 3,
        max_length: int = 256,
    ):
        self.model_name = model_name
        self.max_length = max_length
        self.num_labels = num_labels
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        logger.info("Device: %s", self.device)

        self.tokenizer = AutoTokenizer.from_pretrained(model_name)
        self.encoder = AutoModel.from_pretrained(model_name).to(self.device)

        hidden_size = self.encoder.config.hidden_size
        logger.info("Encoder hidden_size: %d", hidden_size)

        self.classifier = self._build_classifier(hidden_size, num_labels)
        self.loss_fn = nn.CrossEntropyLoss()

        total_params = sum(p.numel() for p in self.parameters())
        logger.info("Total params: %.1fM", total_params / 1e6)
        logger.info("Labels: %s", self.ID2LABEL)

    def _build_classifier(self, hidden_size: int, num_labels: int) -> nn.Sequential:
        return nn.Sequential(
            nn.Dropout(0.1),
            nn.Linear(hidden_size, hidden_size),
            nn.Tanh(),
            nn.Dropout(0.1),
            nn.Linear(hidden_size, num_labels),
        ).to(self.device)

    def _build_inputs(self, texts: list[str], aspects: list[str]) -> dict:
        """
        Формат: [CLS] text [SEP] aspect [SEP]
        tokenizer(text, aspect) автоматически расставляет сегменты.
        """
        return self.tokenizer(
            texts,
            aspects,
            truncation=True,
            padding=True,
            max_length=self.max_length,
            return_tensors="pt",
        ).to(self.device)

    def _encode(self, texts: list[str], aspects: list[str]) -> torch.Tensor:
        """Получаем [CLS] эмбеддинг."""
        inputs = self._build_inputs(texts, aspects)
        outputs = self.encoder(**inputs)
        cls_output = outputs.last_hidden_state[:, 0, :]
        return cls_output

    def forward(
        self,
        texts: list[str],
        aspects: list[str],
        labels: torch.Tensor = None,
    ):
        cls_output = self._encode(texts, aspects)
        logits = self.classifier(cls_output)
        loss = None
        if labels is not None:
            loss = self.loss_fn(logits, labels.to(self.device))
        return _Output(loss=loss, logits=logits)

    def predict(self, text: str, aspect: str) -> dict:
        self.eval_mode()
        with torch.no_grad():
            logits = self.forward([text], [aspect]).logits
        probs = torch.softmax(logits, dim=-1).squeeze()
        pred_id = probs.argmax().item()
        return {
            "aspect": aspect,
            "label": self.ID2LABEL[pred_id],
            "confidence": probs[pred_id].item(),
            "scores": {self.ID2LABEL[i]: p.item() for i, p in enumerate(probs)},
        }

    def predict_batch(
        self, texts: list[str], aspects: list[str], batch_size: int = 32
    ) -> list[dict]:
        self.eval_mode()
        results = []
        with torch.no_grad():
            for i in range(0, len(texts), batch_size):
                bt = texts[i : i + batch_size]
                ba = aspects[i : i + batch_size]
                logits = self.forward(bt, ba).logits
                probs = torch.softmax(logits, dim=-1)
                for j in range(len(bt)):
                    pid = probs[j].argmax().item()
                    results.append(
                        {
                            "aspect": ba[j],
                            "label": self.ID2LABEL[pid],
                            "confidence": probs[j][pid].item(),
                        }
                    )
        return results

    def train_mode(self):
        self.encoder.train()
        self.classifier.train()

    def eval_mode(self):
        self.encoder.eval()
        self.classifier.eval()

    def parameters(self):
        return list(self.encoder.parameters()) + list(self.classifier.parameters())

    def save(self, path: str | Path):
        path = Path(path)
        path.mkdir(parents=True, exist_ok=True)

        logger.info("Saving model to %s", path)

        self.encoder.save_pretrained(path / "encoder")
        self.tokenizer.save_pretrained(path / "encoder")
        torch.save(self.classifier.state_dict(), path / "classifier.pt")

        meta = {
            "num_labels": self.num_labels,
            "max_length": self.max_length,
            "model_name": self.model_name,
            "LABEL2ID": self.LABEL2ID,
            "ID2LABEL": {str(k): v for k, v in self.ID2LABEL.items()},
        }
        with open(path / "meta.json", "w", encoding="utf-8") as f:
            json.dump(meta, f, ensure_ascii=False, indent=2)
        torch.save(meta, path / "config.pt")

        logger.info("Model saved to %s", path)

    def load(self, path: str | Path):
        path = Path(path)
        logger.info("Loading model from %s", path)

        meta = self._load_meta(path)
        self.LABEL2ID = meta["LABEL2ID"]
        self.ID2LABEL = {int(k): v for k, v in meta["ID2LABEL"].items()}
        self.num_labels = meta["num_labels"]
        self.max_length = meta["max_length"]
        self.model_name = meta.get("model_name", self.model_name)

        self.encoder = AutoModel.from_pretrained(path / "encoder").to(self.device)
        self.tokenizer = AutoTokenizer.from_pretrained(path / "encoder")

        hidden_size = self.encoder.config.hidden_size
        self.classifier = self._build_classifier(hidden_size, self.num_labels)
        self.classifier.load_state_dict(
            torch.load(
                path / "classifier.pt",
                map_location=self.device,
                weights_only=True,
            )
        )
        logger.info("Model loaded from %s", path)

    @classmethod
    def from_pretrained(
        cls,
        path: str,
        endpoint_url: str = None,
        aws_access_key_id: str = None,
        aws_secret_access_key: str = None,
        region_name: str = None,
    ) -> AspectSentimentModel:
        """
        path:
          - "models/sentimental/latest"   локальная папка
          - "s3://bucket/prefix"          скачивает из S3 во временную папку
        """
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

        logger.info("Loading meta from %s", local_path)
        meta = cls._load_meta_static(local_path)

        logger.info(
            "Initializing model: model_name=%s, num_labels=%d, max_length=%d",
            meta.get("model_name", "cointegrated/rubert-tiny2"),
            meta["num_labels"],
            meta["max_length"],
        )
        instance = cls(
            model_name=meta.get("model_name", "cointegrated/rubert-tiny2"),
            num_labels=meta["num_labels"],
            max_length=meta["max_length"],
        )

        logger.info("Loading encoder from %s", local_path / "encoder")
        instance.encoder = AutoModel.from_pretrained(
            local_path / "encoder"
        ).to(instance.device)
        instance.tokenizer = AutoTokenizer.from_pretrained(local_path / "encoder")

        instance.LABEL2ID = meta["LABEL2ID"]
        instance.ID2LABEL = {int(k): v for k, v in meta["ID2LABEL"].items()}

        hidden_size = instance.encoder.config.hidden_size
        instance.classifier = instance._build_classifier(
            hidden_size, instance.num_labels
        )

        logger.info("Loading classifier weights from %s", local_path / "classifier.pt")
        instance.classifier.load_state_dict(
            torch.load(
                local_path / "classifier.pt",
                map_location=instance.device,
                weights_only=True,
            )
        )

        total_params = sum(p.numel() for p in instance.parameters())
        logger.info(
            "Model loaded from %s (%.1fM params, device=%s)",
            local_path,
            total_params / 1e6,
            instance.device,
        )
        logger.info("Labels: %s", instance.ID2LABEL)
        return instance

    @staticmethod
    def _download_from_s3(
        s3_path: str,
        endpoint_url: str = None,
        aws_access_key_id: str = None,
        aws_secret_access_key: str = None,
        region_name: str = None,
    ) -> Path:
        """Скачивает модель из S3 во временную локальную папку с прогресс-баром."""
        parts = s3_path.replace("s3://", "").split("/", 1)
        bucket = parts[0]
        prefix = parts[1] if len(parts) > 1 else ""

        logger.info("S3 download: bucket=%s, prefix=%s", bucket, prefix)

        session_kwargs = {}
        if aws_access_key_id:
            session_kwargs["aws_access_key_id"] = aws_access_key_id
        if aws_secret_access_key:
            session_kwargs["aws_secret_access_key"] = aws_secret_access_key
        if region_name:
            session_kwargs["region_name"] = region_name

        session = boto3.session.Session(**session_kwargs)
        client_kwargs = {}
        if endpoint_url:
            client_kwargs["endpoint_url"] = endpoint_url
        s3 = session.client("s3", **client_kwargs)

        logger.info("Listing objects at s3://%s/%s ...", bucket, prefix)
        all_objects = []
        continuation_token = None

        while True:
            list_kwargs = {"Bucket": bucket, "Prefix": prefix}
            if continuation_token:
                list_kwargs["ContinuationToken"] = continuation_token
            response = s3.list_objects_v2(**list_kwargs)

            if "Contents" not in response:
                break
            all_objects.extend(response["Contents"])

            if response.get("IsTruncated"):
                continuation_token = response["NextContinuationToken"]
            else:
                break

        if not all_objects:
            raise FileNotFoundError(f"No objects found at s3://{bucket}/{prefix}")

        files_to_download = []
        total_bytes = 0
        for obj in all_objects:
            key = obj["Key"]
            rel = key[len(prefix):].lstrip("/")
            if not rel:
                continue
            files_to_download.append((key, rel, obj["Size"]))
            total_bytes += obj["Size"]

        logger.info(
            "Found %d files (%.2f MB total)",
            len(files_to_download),
            total_bytes / 1e6,
        )

        local_dir = Path(tempfile.mkdtemp(prefix="absa_model_"))
        logger.info("Downloading to %s", local_dir)

        downloaded_bytes = 0

        if tqdm is not None:
            progress = tqdm(
                total=total_bytes,
                unit="B",
                unit_scale=True,
                unit_divisor=1024,
                desc="Downloading model from S3",
                ncols=100,
            )
        else:
            progress = None

        for key, rel, size in files_to_download:
            local_file = local_dir / rel
            local_file.parent.mkdir(parents=True, exist_ok=True)

            logger.debug("  ↓ s3://%s/%s  %s (%s bytes)", bucket, key, local_file, size)

            if progress is not None:
                def _make_callback(pbar):
                    def _cb(bytes_transferred):
                        pbar.update(bytes_transferred)
                    return _cb

                s3.download_file(
                    bucket,
                    key,
                    str(local_file),
                    Callback=_make_callback(progress),
                )
            else:
                s3.download_file(bucket, key, str(local_file))
                downloaded_bytes += size
                pct = downloaded_bytes / total_bytes * 100 if total_bytes else 0
                logger.info(
                    "  [%3.0f%%] %s (%.1f KB)",
                    pct,
                    rel,
                    size / 1024,
                )

        if progress is not None:
            progress.close()

        logger.info("S3 download complete  %s", local_dir)
        return local_dir

    @staticmethod
    def _load_meta_static(path: Path) -> dict:
        """Загрузка мета-конфига (поддерживает meta.json и config.pt)."""
        meta_json = path / "meta.json"
        config_pt = path / "config.pt"

        if meta_json.exists():
            logger.debug("Loading meta from %s", meta_json)
            with open(meta_json, encoding="utf-8") as f:
                meta = json.load(f)
        elif config_pt.exists():
            logger.debug("Loading meta from %s", config_pt)
            meta = torch.load(config_pt, weights_only=False)
        else:
            raise FileNotFoundError(
                f"Neither meta.json nor config.pt found in {path}"
            )

        if "ID2LABEL" in meta:
            meta["ID2LABEL"] = {int(k): v for k, v in meta["ID2LABEL"].items()}

        logger.debug("Meta loaded: %s", meta)
        return meta

    def _load_meta(self, path: Path) -> dict:
        return self._load_meta_static(path)

    @staticmethod
    def load_production() -> AspectSentimentModel:
        try:
            config_path = Path("src/models/sentimental/configs/prod_config.yml")
            logger.info("Loading production config from %s", config_path)

            with open(config_path) as f:
                cfg = yaml.safe_load(f)

            model_path = cfg["model"]["name"]
            logger.info("Production model path: %s", model_path)

            cfg = S3Config()
            return AspectSentimentModel.from_pretrained(
                path=model_path,
                endpoint_url=cfg.S3_ENDPOINT_URL,
                aws_access_key_id=cfg.S3_ACCESS_KEY,
                aws_secret_access_key=cfg.S3_SECRET_KEY,
                region_name=cfg.S3_REGION_NAME
            )
        except Exception as e:
            logger.error("Failed to load production model: %s", str(e))
            raise e


class _Output:
    """Простая обёртка для совместимости с HF-style output."""

    def __init__(self, loss, logits):
        self.loss = loss
        self.logits = logits
