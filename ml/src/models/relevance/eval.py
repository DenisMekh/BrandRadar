from __future__ import annotations

from pathlib import Path

from model import CompanyRelevanceModel
from dev import load_config, prepare_data, evaluate
from train import compute_metrics


def eval_model(cfg: dict):
    model_path = cfg["model"]["name"]

    print(f"Loading model from: {model_path}")
    rm = CompanyRelevanceModel.from_pretrained(str(model_path))

    _, test_df = prepare_data(cfg, rm, split=True)
    print(f"Test samples: {len(test_df)}")

    acc, y_true, y_pred, _ = evaluate(rm, test_df, verbose=True)
    metrics = compute_metrics(y_true, y_pred)

    print(f"\n{'═'*40}")
    print("ИТОГО")
    print(f"{'═'*40}")
    print(f"  Accuracy:   {metrics['accuracy']:.4f}")
    print(f"  F1:         {metrics['f1']:.4f}")
    print(f"  Precision:  {metrics['precision']:.4f}")
    print(f"  Recall:     {metrics['recall']:.4f}")
    print(f"{'═'*40}")

    return metrics


if __name__ == "__main__":
    cfg = load_config()
    eval_model(cfg)