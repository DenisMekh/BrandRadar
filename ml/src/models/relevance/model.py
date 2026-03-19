from __future__ import annotations

import sys
import json
import logging
import torch
import torch.nn as nn
from transformers import AutoTokenizer, AutoModel
from pathlib import Path
import yaml

sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent))
from src.s3config import S3Config

logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)


class CompanyRelevanceModel:
    """
    Бинарная классификация: релевантна ли компания тексту сообщения.

    Вход:  текст сообщения + название компании + ключевые слова
    Выход: relevant / irrelevant
    
    Поддерживает два режима работы:
    1. Базовый: text + company (для обратной совместимости)
    2. Расширенный: text + company + keywords (новый формат)
    """

    LABEL2ID = {
        "irrelevant": 0,
        "relevant": 1,
    }
    ID2LABEL = {v: k for k, v in LABEL2ID.items()}

    def __init__(
        self,
        model_name: str = "cointegrated/rubert-tiny2",
        num_labels: int = 2,
        max_length: int = 256,
        use_keywords: bool = True,
    ):
        self.model_name = model_name
        self.max_length = max_length
        self.num_labels = num_labels
        self.use_keywords = use_keywords
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        logger.info("Device: %s", self.device)
        logger.info("Use keywords: %s", self.use_keywords)

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

    def _build_inputs(
        self,
        texts: list[str],
        companies: list[str],
        keywords: list[str] | None = None,
    ) -> dict:
        """
        Строит входные данные для модели.
        
        Если keywords предоставлены, они добавляются к тексту компании для лучшего контекста.
        Формат: [CLS] текст [SEP] компания: ключевые слова [SEP]
        """
        if self.use_keywords and keywords is not None:
            company_contexts = [
                f"{company}: {kw}" if kw else company
                for company, kw in zip(companies, keywords)
            ]
            return self.tokenizer(
                texts,
                company_contexts,
                truncation=True,
                padding=True,
                max_length=self.max_length,
                return_tensors="pt",
            ).to(self.device)
        else:
            return self.tokenizer(
                texts,
                companies,
                truncation=True,
                padding=True,
                max_length=self.max_length,
                return_tensors="pt",
            ).to(self.device)

    def _encode(
        self,
        texts: list[str],
        companies: list[str],
        keywords: list[str] | None = None,
    ) -> torch.Tensor:
        inputs = self._build_inputs(texts, companies, keywords)
        outputs = self.encoder(**inputs)
        cls_output = outputs.last_hidden_state[:, 0, :]
        return cls_output

    def forward(
        self,
        texts: list[str],
        companies: list[str],
        keywords: list[str] | None = None,
        labels: torch.Tensor = None,
    ):
        cls_output = self._encode(texts, companies, keywords)
        logits = self.classifier(cls_output)

        loss = None
        if labels is not None:
            loss = self.loss_fn(logits, labels.to(self.device))

        return _Output(loss=loss, logits=logits)

    def predict(
        self,
        text: str,
        company: str,
        keywords: str | None = None,
    ) -> dict:
        """
        Предсказание релевантности для одного примера.
        
        Args:
            text: Текст сообщения
            company: Название компании
            keywords: Ключевые слова компании (опционально, для нового формата)
        
        Returns:
            dict с результатами предсказания
        """
        self.eval_mode()
        with torch.no_grad():
            logits = self.forward([text], [company], keywords=[keywords] if keywords else None).logits

        probs = torch.softmax(logits, dim=-1).squeeze()
        pred_id = probs.argmax().item()

        return {
            "company": company,
            "label": self.ID2LABEL[pred_id],
            "is_relevant": pred_id == 1,
            "confidence": probs[pred_id].item(),
            "scores": {self.ID2LABEL[i]: p.item() for i, p in enumerate(probs)},
        }

    def predict_batch(
        self,
        texts: list[str],
        companies: list[str],
        keywords: list[str] | None = None,
        batch_size: int = 32,
    ) -> list[dict]:
        """
        Пакетное предсказание релевантности.
        
        Args:
            texts: Список текстов сообщений
            companies: Список названий компаний
            keywords: Список ключевых слов (опционально, для нового формата)
            batch_size: Размер пакета
        
        Returns:
            Список dict с результатами предсказания
        """
        self.eval_mode()
        results = []

        with torch.no_grad():
            for i in range(0, len(texts), batch_size):
                bt = texts[i : i + batch_size]
                bc = companies[i : i + batch_size]
                bk = keywords[i : i + batch_size] if keywords is not None else None
                logits = self.forward(bt, bc, keywords=bk).logits
                probs = torch.softmax(logits, dim=-1)

                for j in range(len(bt)):
                    pid = probs[j].argmax().item()
                    results.append(
                        {
                            "company": bc[j],
                            "label": self.ID2LABEL[pid],
                            "is_relevant": pid == 1,
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
            "model_type": "company_relevance",
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
    ) -> CompanyRelevanceModel:
        cfg = S3Config()
        if path.startswith("s3://"):
            logger.info("Detected S3 path: %s", path)
            local_path = cls._download_from_s3(
                s3_path=path,
                endpoint_url=cfg.S3_ENDPOINT_URL,
                aws_access_key_id=cfg.S3_ACCESS_KEY,
                aws_secret_access_key=cfg.S3_SECRET_KEY,
                region_name=cfg.S3_REGION_NAME,
            )
        else:
            local_path = Path(path)
            if local_path.is_symlink():
                local_path = local_path.resolve()
            logger.info("Using local path: %s", local_path)

        if not local_path.exists():
            raise FileNotFoundError(f"Model path not found: {local_path}")

        meta = cls._load_meta_static(local_path)

        instance = cls(
            model_name=meta.get("model_name", "cointegrated/rubert-tiny2"),
            num_labels=meta["num_labels"],
            max_length=meta["max_length"],
        )

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
        return instance

    @staticmethod
    def _download_from_s3(
        s3_path: str,
        endpoint_url: str = None,
        aws_access_key_id: str = None,
        aws_secret_access_key: str = None,
        region_name: str = None,
    ) -> Path:
        import tempfile

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
        client_kwargs = {}
        if endpoint_url:
            client_kwargs["endpoint_url"] = endpoint_url
        s3 = session.client("s3", **client_kwargs)

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

        local_dir = Path(tempfile.mkdtemp(prefix="relevance_model_"))

        if tqdm is not None:
            progress = tqdm(total=total_bytes, unit="B", unit_scale=True, unit_divisor=1024)
        else:
            progress = None

        for key, rel, size in files_to_download:
            local_file = local_dir / rel
            local_file.parent.mkdir(parents=True, exist_ok=True)
            if progress is not None:
                def _make_callback(pbar):
                    def _cb(bytes_transferred):
                        pbar.update(bytes_transferred)
                    return _cb
                s3.download_file(bucket, key, str(local_file), Callback=_make_callback(progress))
            else:
                s3.download_file(bucket, key, str(local_file))

        if progress is not None:
            progress.close()

        return local_dir

    @staticmethod
    def _load_meta_static(path: Path) -> dict:
        meta_json = path / "meta.json"
        config_pt = path / "config.pt"

        if meta_json.exists():
            with open(meta_json, encoding="utf-8") as f:
                meta = json.load(f)
        elif config_pt.exists():
            meta = torch.load(config_pt, weights_only=False)
        else:
            raise FileNotFoundError(f"Neither meta.json nor config.pt found in {path}")

        if "ID2LABEL" in meta:
            meta["ID2LABEL"] = {int(k): v for k, v in meta["ID2LABEL"].items()}

        return meta

    def _load_meta(self, path: Path) -> dict:
        return self._load_meta_static(path)

    @staticmethod
    def load_production() -> CompanyRelevanceModel:
        config_path = Path("src/models/relevance/configs/prod_config.yml")
        with open(config_path) as f:
            model_config = yaml.safe_load(f)
        model_path = model_config["model"]["name"]


        cfg = S3Config()

        return CompanyRelevanceModel.from_pretrained(
            path=model_path,
            endpoint_url=cfg.S3_ENDPOINT_URL,
            aws_access_key_id=cfg.S3_ACCESS_KEY,
            aws_secret_access_key=cfg.S3_SECRET_KEY,
            region_name=cfg.S3_REGION_NAME
        )

    def freeze_encoder(self):
        for param in self.encoder.parameters():
            param.requires_grad = False

    def unfreeze_encoder(self, from_layer: int | None = None):
        if from_layer is None:
            for param in self.encoder.parameters():
                param.requires_grad = True
        else:
            for param in self.encoder.parameters():
                param.requires_grad = False
            for i, layer in enumerate(self.encoder.encoder.layer):
                if i >= from_layer:
                    for param in layer.parameters():
                        param.requires_grad = True
            if hasattr(self.encoder, "pooler") and self.encoder.pooler is not None:
                for param in self.encoder.pooler.parameters():
                    param.requires_grad = True

    def trainable_parameters(self) -> list[torch.nn.Parameter]:
        return [p for p in self.parameters() if p.requires_grad]


    def forward_encoded(
            self,
            input_ids: torch.Tensor,
            attention_mask: torch.Tensor,
            token_type_ids: torch.Tensor | None = None,
            labels: torch.Tensor | None = None,
    ) -> _Output:
        kwargs = {
            "input_ids": input_ids.to(self.device),
            "attention_mask": attention_mask.to(self.device),
        }
        if token_type_ids is not None:
            kwargs["token_type_ids"] = token_type_ids.to(self.device)

        outputs = self.encoder(**kwargs)
        cls_output = outputs.last_hidden_state[:, 0, :]
        logits = self.classifier(cls_output)

        loss = None
        if labels is not None:
            loss = self.loss_fn(logits, labels.to(self.device))
        return _Output(loss=loss, logits=logits)


class _Output:
    def __init__(self, loss, logits):
        self.loss = loss
        self.logits = logits
