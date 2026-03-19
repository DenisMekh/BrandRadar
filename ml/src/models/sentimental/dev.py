import sys
from pathlib import Path

import yaml
import torch
import random
import numpy as np
import pandas as pd
from sklearn.metrics import classification_report, accuracy_score, confusion_matrix
from sklearn.model_selection import train_test_split
from torch.utils.data import DataLoader, Dataset
from tqdm import tqdm

from model import AspectSentimentModel

sys.path.insert(0, str(Path(__file__).parent.parent.parent.parent))
from src.s3config import S3Config

CONFIG_PATH = "src/models/sentimental/configs/dev_config.yml"


def set_seed(seed: int):
    """Set random seeds for reproducibility."""
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    torch.cuda.manual_seed_all(seed)


def load_config(path: str = CONFIG_PATH) -> dict:
    """Load configuration from YAML file."""
    with open(path) as f:
        return yaml.safe_load(f)


def create_or_load_model(model_path: str) -> AspectSentimentModel:
    """
    If model.name starts with 'models/' or 's3://' — load pretrained model via from_pretrained.
    Otherwise — create new model from HuggingFace.
    """



    cfg = S3Config()

    if model_path.startswith("models/") or model_path.startswith("s3://"):
        print(f"Loading pretrained model from: {model_path}")
        sm = AspectSentimentModel.from_pretrained(
            path=model_path,
            endpoint_url=cfg.S3_ENDPOINT_URL,
            aws_access_key_id=cfg.S3_ACCESS_KEY,
            aws_secret_access_key=cfg.S3_SECRET_KEY,
            region_name=cfg.S3_REGION_NAME
        )
        print(f"Pretrained model loaded. Labels: {sm.ID2LABEL}")
    else:
        print(f"Creating new model from HuggingFace: {model_path}")
        sm = AspectSentimentModel(
            model_name=model_path,
            num_labels=cfg["model"]["num_labels"],
            max_length=cfg["model"]["max_length"],
        )

    return sm


class AspectDataset(Dataset):
    """Dataset for aspect sentiment analysis."""

    def __init__(self, texts: list[str], aspects: list[str], labels: list[int]):
        self.texts = texts
        self.aspects = aspects
        self.labels = labels

    def __len__(self):
        return len(self.texts)

    def __getitem__(self, idx):
        return self.texts[idx], self.aspects[idx], self.labels[idx]


def collate_fn(batch):
    """Collate function for DataLoader."""
    texts, aspects, labels = zip(*batch)
    return list(texts), list(aspects), torch.tensor(list(labels))


def prepare_data(cfg: dict, sm: AspectSentimentModel, split: bool = False):
    """
    Prepare data for training/evaluation.

    Supports two dataset formats:
    1. Old format: "title", "text", "company", "sentiment"
    2. New format: "text", "brand", "sentiment"

    Args:
        cfg: Configuration dict
        sm: AspectSentimentModel instance
        split: If True, return train/test split (for training).
               If False, return single dataframe (for evaluation).

    Returns:
        If split=True: (train_df, test_df)
        If split=False: df
    """
    df = pd.read_csv(cfg["data"]["path"])

    print("\n" + "=" * 60)
    print("ПЕРВЫЕ СТРОКИ ДАТАСЕТА")
    print("=" * 60)
    print(df.head(10).to_string())
    print(f"\nВсего строк: {len(df)}")
    print(f"Колонки: {list(df.columns)}")
    print("=" * 60 + "\n")

    if "brand" in df.columns and "company" not in df.columns:
        print("ℹ️  Обнаружен новый формат датасета: используем 'brand' как 'company'")
        df["company"] = df["brand"]
    
    if "company" not in df.columns:
        print("⚠️  Колонка 'company' не найдена — используем заглушку 'компания'")
        df["company"] = "компания"

    df["label_id"] = df["sentiment"].map(sm.LABEL2ID)
    df = df.dropna(subset=["label_id"])
    df["label_id"] = df["label_id"].astype(int)

    if "title" not in df.columns:
        print("ℹ️  Колонка 'title' не найдена — используем только 'text'")
        df["title"] = ""
    
    df["title"] = df["title"].fillna("").astype(str)
    df["text"] = df["text"].fillna("").astype(str)
    df["input"] = (df["title"] + " " + df["text"]).str.strip()
    df = df[df["input"].str.len() > 0]

    sample_size = cfg["data"].get("sample_size")
    if sample_size and sample_size < len(df):
        df = df.groupby("label_id", group_keys=False).apply(
            lambda x: x.sample(
                n=min(len(x), sample_size // df["label_id"].nunique()),
                random_state=cfg["train"]["seed"],
            )
        )

    if split:
        train_df, test_df = train_test_split(
            df,
            test_size=cfg["data"]["test_size"],
            random_state=cfg["train"]["seed"],
            stratify=df["label_id"],
        )
        return train_df, test_df

    return df


def evaluate(sm: AspectSentimentModel, df: pd.DataFrame, verbose: bool = True):
    """
    Evaluate model on dataframe.

    Args:
        sm: AspectSentimentModel instance
        df: DataFrame with 'input', 'company', 'label_id' columns
        verbose: If True, print detailed reports

    Returns:
        acc, y_true, y_pred, all_probs
    """
    sm.eval_mode()
    texts = df["input"].tolist()
    aspects = df["company"].tolist()
    y_true = df["label_id"].tolist()
    y_pred = []
    all_probs = []

    with torch.no_grad():
        for i in tqdm(range(0, len(texts), 32), desc="Оценка модели", unit="batch"):
            bt = texts[i : i + 32]
            ba = aspects[i : i + 32]
            logits = sm.forward(bt, ba).logits
            probs = torch.softmax(logits, dim=-1)
            y_pred.extend(logits.argmax(dim=-1).cpu().tolist())
            all_probs.extend(probs.cpu().tolist())

    acc = accuracy_score(y_true, y_pred)

    if verbose:
        labels = sorted(sm.ID2LABEL.keys())
        names = [sm.ID2LABEL[i] for i in labels]

        print("\n" + "=" * 60)
        print("ОЦЕНКА НА ДАТАСЕТЕ")
        print("=" * 60)
        print(f"\nAccuracy: {acc:.4f}\n")
        print(classification_report(y_true, y_pred, target_names=names))

        cm = confusion_matrix(y_true, y_pred, labels=labels)
        print("Матрица ошибок:")
        header = "           " + "  ".join(f"{n:>10}" for n in names)
        print(header)
        for i, row in enumerate(cm):
            row_str = "  ".join(f"{v:>10}" for v in row)
            print(f"  {names[i]:>10} | {row_str}")

        probs_arr = np.array(all_probs)
        max_probs = probs_arr.max(axis=1)
        print(f"\nУверенность модели:")
        print(f"  Средняя:  {max_probs.mean():.3f}")
        print(f"  <0.5:     {(max_probs < 0.5).sum()} ({(max_probs < 0.5).mean():.1%})")
        print(f"  >0.9:     {(max_probs > 0.9).sum()} ({(max_probs > 0.9).mean():.1%})")

    return acc, y_true, y_pred, all_probs


def show_errors(sm: AspectSentimentModel, df: pd.DataFrame, y_true, y_pred, n=10):
    """Display error examples from predictions."""
    print("\n" + "=" * 60)
    print("ПРИМЕРЫ ОШИБОК")
    print("=" * 60)

    df = df.copy()
    df["y_true"] = y_true
    df["y_pred"] = y_pred
    df["true_label"] = df["y_true"].map(sm.ID2LABEL)
    df["pred_label"] = df["y_pred"].map(sm.ID2LABEL)

    errors = df[df["y_true"] != df["y_pred"]]
    print(f"\nОшибок: {len(errors)} из {len(df)} ({len(errors)/len(df):.1%})")

    print("\nТоп путаниц:")
    pairs = (
        errors.groupby(["true_label", "pred_label"]).size().sort_values(ascending=False)
    )
    for (t, p), cnt in pairs.head(6).items():
        print(f"  {t:>10}  {p:<10}: {cnt}")

    print(f"\nПримеры ошибок (топ-{n}):")
    for _, row in errors.head(n).iterrows():
        text = row["input"][:100]
        print(
            f"  аспект={row['company']:<15} "
            f"ожидали={row['true_label']:<10} "
            f"предсказано={row['pred_label']:<10} | "
            f"{text}{'...' if len(row['input']) > 100 else ''}"
        )


def test_contrastive(sm: AspectSentimentModel):
    """Test model on contrastive examples (same text, different aspects)."""
    print("\n" + "=" * 60)
    print("ТЕСТ: КОНТРАСТНЫЕ ПРИМЕРЫ (одна и та же фраза, разные аспекты)")
    print("=" * 60)

    cases = [
        {
            "text": "В Тинькофф ужасное обслуживание, а вот в Сбербанке всегда помогут",
            "checks": [
                ("Сбербанк", "positive"),
                ("Тинькофф", "negative"),
            ],
        },
        {
            "text": "Озон доставляет быстро, Wildberries постоянно задерживает",
            "checks": [
                ("Озон", "positive"),
                ("Wildberries", "negative"),
            ],
        },
        {
            "text": "МТС и Билайн одинаково плохо ловят в метро",
            "checks": [
                ("МТС", "negative"),
                ("Билайн", "negative"),
            ],
        },
        {
            "text": "Яндекс Такси нормально работает, ничего особенного",
            "checks": [
                ("Яндекс Такси", "neutral"),
            ],
        },
    ]

    correct = 0
    total = 0

    for case in cases:
        text = case["text"]
        print(f"\n  Текст: {text}")

        for aspect, expected in case["checks"]:
            result = sm.predict(text, aspect)
            pred = result["label"]
            conf = result["confidence"]
            match = "✅" if pred == expected else "❌"

            if pred == expected:
                correct += 1
            total += 1

            print(
                f"    {match} аспект={aspect:<15} ожидали={expected:<10} "
                f"предсказано={pred:<10} ({conf:.2f})"
            )

    print(f"\nКонтрастные: {correct}/{total} верно")


def test_simple(sm: AspectSentimentModel):
    """Test model on simple examples."""
    print("\n" + "=" * 60)
    print("ТЕСТ: ПРОСТЫЕ ПРИМЕРЫ")
    print("=" * 60)

    examples = [
        ("Отличный банк, очень доволен", "Сбербанк", "positive"),
        ("Ужасный сервис, больше не приду", "Тинькофф", "negative"),
        ("Нормально работает, без претензий", "Озон", "neutral"),
        ("Лучший маркетплейс, рекомендую всем", "Wildberries", "positive"),
        ("Постоянные сбои, приложение не работает", "МТС", "negative"),
        ("Доставка стандартная, ничего необычного", "СДЭК", "neutral"),
    ]

    correct = 0
    for text, aspect, expected in examples:
        result = sm.predict(text, aspect)
        pred = result["label"]
        conf = result["confidence"]
        match = "✅" if pred == expected else "❌"

        if pred == expected:
            correct += 1

        print(
            f"  {match} аспект={aspect:<15} ожидали={expected:<10} "
            f"предсказано={pred:<10} ({conf:.2f}) | {text}"
        )

    print(f"\nПростые: {correct}/{len(examples)} верно")
