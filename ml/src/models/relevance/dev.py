import yaml
import torch
import random
import numpy as np
import pandas as pd
from sklearn.metrics import (
    classification_report,
    accuracy_score,
    confusion_matrix,
    f1_score,
    precision_score,
    recall_score,
)
from sklearn.model_selection import train_test_split
from torch.utils.data import DataLoader, Dataset

try:
    from tqdm import tqdm
except ImportError:
    tqdm = None

from model import CompanyRelevanceModel

CONFIG_PATH = "src/models/relevance/configs/dev_config.yml"


def set_seed(seed: int):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)
    torch.cuda.manual_seed_all(seed)


def load_config(path: str = CONFIG_PATH) -> dict:
    with open(path) as f:
        return yaml.safe_load(f)


def create_or_load_model(cfg: dict) -> CompanyRelevanceModel:
    model_name = cfg["model"]["name"]
    use_keywords = cfg["model"].get("use_keywords", True)

    if model_name.startswith("models/") or model_name.startswith("s3://"):
        print(f"Loading pretrained model from: {model_name}")
        rm = CompanyRelevanceModel.from_pretrained(
            path=model_name,
            endpoint_url=cfg.get("s3", {}).get("endpoint_url"),
            aws_access_key_id=cfg.get("s3", {}).get("aws_access_key_id"),
            aws_secret_access_key=cfg.get("s3", {}).get("aws_secret_access_key"),
            region_name=cfg.get("s3", {}).get("region_name"),
        )
        print(f"Pretrained model loaded. Labels: {rm.ID2LABEL}")
    else:
        print(f"Creating new model from HuggingFace: {model_name}")
        rm = CompanyRelevanceModel(
            model_name=model_name,
            num_labels=cfg["model"]["num_labels"],
            max_length=cfg["model"]["max_length"],
            use_keywords=use_keywords,
        )

    return rm


class RelevanceDataset(Dataset):
    def __init__(self, texts: list[str], companies: list[str], labels: list[int]):
        self.texts = texts
        self.companies = companies
        self.labels = labels

    def __len__(self):
        return len(self.texts)

    def __getitem__(self, idx):
        return self.texts[idx], self.companies[idx], self.labels[idx]


def collate_fn(batch):
    texts, companies, labels = zip(*batch)
    return list(texts), list(companies), torch.tensor(list(labels))


def prepare_data(cfg: dict, rm: CompanyRelevanceModel, split: bool = False):
    """
    Ожидаемые колонки в CSV (новый формат):
      - text      — текст сообщения
      - company   — название компании
      - keywords  — ключевые слова компании (через точку с запятой)
      - relevant  — метка: 1 (релевантно) / 0 (нерелевантно)
    
    Старый формат также поддерживается:
      - title     — заголовок (может быть пустым)
      - text      — текст сообщения
      - label     — метка: irrelevant / relevant
      - company   — название компании
    """
    df = pd.read_csv(cfg["data"]["path"])

    print(f"\nЗагружено строк: {len(df)}")
    print(f"Колонки: {list(df.columns)}")

    is_new_format = "relevant" in df.columns and "keywords" in df.columns
    
    if is_new_format:
        print("Обнаружен новый формат датасета (text, company, keywords, relevant)")
        
        if "company" not in df.columns:
            raise ValueError("Колонка 'company' не найдена в новом формате датасета")
        if "text" not in df.columns:
            raise ValueError("Колонка 'text' не найдена в новом формате датасета")
        if "keywords" not in df.columns:
            raise ValueError("Колонка 'keywords' не найдена в новом формате датасета")

        df["label_id"] = df["relevant"].astype(int)

        df["text"] = df["text"].fillna("").astype(str).str.strip()

        df["keywords"] = df["keywords"].fillna("").astype(str).str.strip()

        df["input"] = df["text"]
        
        before = len(df)
        df = df[df["input"].str.len() > 0]
        dropped = before - len(df)
        if dropped > 0:
            print(f"⚠️  Удалено {dropped} строк с пустым текстом")
    else:
        print("Обнаружен старый формат датасета (title, text, label, company)")
        
        if "company" not in df.columns:
            print("⚠️  Колонка 'company' не найдена — используем заглушку 'компания'")
            df["company"] = "компания"

        label_col = None
        for col in ["label", "relevance"]:
            if col in df.columns:
                label_col = col
                break

        if label_col is None:
            raise ValueError(
                "Не найдена колонка с метками. "
                "Ожидается 'label' или 'relevance' со значениями: irrelevant, relevant"
            )

        print(f"Колонка с метками: '{label_col}'")
        print(f"Уникальные метки: {df[label_col].unique().tolist()}")

        df["label_id"] = df[label_col].map(rm.LABEL2ID)

        unmapped = df[df["label_id"].isna()][label_col].unique()
        if len(unmapped) > 0:
            print(f"⚠️  Неизвестные метки (будут удалены): {unmapped.tolist()}")

        df = df.dropna(subset=["label_id"])
        df["label_id"] = df["label_id"].astype(int)

        df["title"] = df["title"].fillna("").astype(str) if "title" in df.columns else ""
        df["text"] = df["text"].fillna("").astype(str) if "text" in df.columns else ""

        if "title" in df.columns:
            df["input"] = (df["title"].str.strip() + " " + df["text"].str.strip()).str.strip()
        else:
            df["input"] = df["text"].str.strip()

        before = len(df)
        df = df[df["input"].str.len() > 0]
        dropped = before - len(df)
        if dropped > 0:
            print(f"⚠️  Удалено {dropped} строк с пустым текстом")

    sample_size = cfg["data"].get("sample_size")
    if sample_size and sample_size < len(df):
        df = df.groupby("label_id", group_keys=False).apply(
            lambda x: x.sample(
                n=min(len(x), sample_size // df["label_id"].nunique()),
                random_state=cfg["train"]["seed"],
            )
        )

    print(f"\nРаспределение меток:")
    for label_id in sorted(rm.ID2LABEL.keys()):
        label_name = rm.ID2LABEL[label_id]
        count = (df["label_id"] == label_id).sum()
        pct = count / len(df) * 100 if len(df) > 0 else 0
        print(f"  {label_name:<15}: {count:>5} ({pct:.1f}%)")
    print(f"  {'TOTAL':<15}: {len(df):>5}")

    print(f"\nКомпании в данных:")
    for company, count in df["company"].value_counts().head(10).items():
        print(f"  {company:<20}: {count:>5}")
    if df["company"].nunique() > 10:
        print(f"  ... и ещё {df['company'].nunique() - 10} компаний")

    if split:
        train_df, test_df = train_test_split(
            df,
            test_size=cfg["data"]["test_size"],
            random_state=cfg["train"]["seed"],
            stratify=df["label_id"],
        )
        return train_df, test_df

    return df


def evaluate(rm: CompanyRelevanceModel, df: pd.DataFrame, verbose: bool = True):
    """
    Бинарная оценка: irrelevant (0) vs relevant (1).
    
    Поддерживает оба формата:
    - Старый: input + company
    - Новый: input + company + keywords
    """
    rm.eval_mode()

    texts = df["input"].tolist()
    companies = df["company"].tolist()
    y_true = df["label_id"].tolist()

    keywords = df["keywords"].tolist() if "keywords" in df.columns else None

    y_pred = []
    all_probs = []

    batch_indices = list(range(0, len(texts), 32))
    iterator = tqdm(batch_indices, desc="Evaluating", unit="batch") if tqdm else batch_indices

    with torch.no_grad():
        for i in iterator:
            bt = texts[i : i + 32]
            bc = companies[i : i + 32]
            bk = keywords[i : i + 32] if keywords is not None else None
            logits = rm.forward(bt, bc, keywords=bk).logits
            probs = torch.softmax(logits, dim=-1)
            y_pred.extend(logits.argmax(dim=-1).cpu().tolist())
            all_probs.extend(probs.cpu().tolist())

    acc = accuracy_score(y_true, y_pred)
    f1 = f1_score(y_true, y_pred, pos_label=1)
    prec = precision_score(y_true, y_pred, pos_label=1)
    rec = recall_score(y_true, y_pred, pos_label=1)

    if verbose:
        names = ["irrelevant", "relevant"]

        print("\n" + "=" * 60)
        print("ОЦЕНКА НА ДАТАСЕТЕ (БИНАРНАЯ РЕЛЕВАНТНОСТЬ)")
        print("=" * 60)

        print(f"\n  Accuracy:  {acc:.4f}")
        print(f"  F1:        {f1:.4f}")
        print(f"  Precision: {prec:.4f}")
        print(f"  Recall:    {rec:.4f}\n")

        print(classification_report(y_true, y_pred, target_names=names))

        cm = confusion_matrix(y_true, y_pred, labels=[0, 1])
        print("Матрица ошибок:")
        print(f"  {'':>15} {'irrelevant':>12} {'relevant':>12}")
        for i, row in enumerate(cm):
            row_str = "  ".join(f"{v:>10}" for v in row)
            print(f"  {names[i]:>15} | {row_str}")

        tn, fp, fn, tp = cm.ravel()
        print(f"\n  TP={tp}  FP={fp}  FN={fn}  TN={tn}")
        print(f"  False positives (irrelevant  relevant): {fp}")
        print(f"  False negatives (relevant  irrelevant): {fn}")

        probs_arr = np.array(all_probs)
        max_probs = probs_arr.max(axis=1)
        print(f"\nУверенность модели:")
        print(f"  Средняя:  {max_probs.mean():.3f}")
        print(f"  Медиана:  {np.median(max_probs):.3f}")
        print(f"  <0.6:     {(max_probs < 0.6).sum()} ({(max_probs < 0.6).mean():.1%})")
        print(f"  >0.9:     {(max_probs > 0.9).sum()} ({(max_probs > 0.9).mean():.1%})")

    return acc, y_true, y_pred, all_probs
