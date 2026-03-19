import pandas as pd
from pathlib import Path

RAW_PATH = Path("data/raw.xlsx")
CLEAN_PATH = Path("data/clean/sentiment.csv")

LABEL_MAP = {
    "нейтрально": "neutral",
    "позитив": "positive",
    "негатив": "negative",
}


df = pd.read_excel(RAW_PATH)
df = df[["Заголовок", "Текст", "Тональность"]].copy()

df["Заголовок"] = df["Заголовок"].fillna("").str.strip()
df["Текст"] = df["Текст"].fillna("").str.strip()
df["Тональность"] = df["Тональность"].fillna("unknown").str.strip().str.lower()

df = df[(df["Заголовок"] != "") | (df["Текст"] != "")]

df["Тональность"] = df["Тональность"].map(LABEL_MAP).fillna("unknown")

df = df[df["Тональность"] != "unknown"]

df.columns = ["title", "text", "sentiment"]

df = df.drop_duplicates(subset=["title", "text"])
df['company'] = 'Т-Банк'

CLEAN_PATH.parent.mkdir(parents=True, exist_ok=True)
df.to_csv(CLEAN_PATH, index=False)

print(f"Rows: {len(df)}, Labels: {df['sentiment'].value_counts().to_dict()}")