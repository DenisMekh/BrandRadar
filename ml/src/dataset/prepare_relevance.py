import pandas as pd
from pathlib import Path

RAW_PATH = Path("data/raw.xlsx")
CLEAN_PATH = Path("data/clean/relevance.csv")

LABEL_MAP = {
    "релевант": "relevant",
    "нерелевант": "irrelevant",
}

df = pd.read_excel(RAW_PATH)
df = df[["Заголовок", "Текст", "Редевантность"]].copy()

df["Заголовок"] = df["Заголовок"].fillna("").str.strip()
df["Текст"] = df["Текст"].fillna("").str.strip()
df["Редевантность"] = df["Редевантность"].fillna("unknown").str.strip().str.lower()

df = df[(df["Заголовок"] != "") | (df["Текст"] != "")]

df["Редевантность"] = df["Редевантность"].map(LABEL_MAP).fillna("unknown")
df = df[df["Редевантность"] != "unknown"]

df.columns = ["title", "text", "label"]
df = df.drop_duplicates(subset=["title", "text"])
df["company"] = "Т-Банк"

CLEAN_PATH.parent.mkdir(parents=True, exist_ok=True)
df.to_csv(CLEAN_PATH, index=False)

print(f"Rows: {len(df)}, Labels: {df['label'].value_counts().to_dict()}")