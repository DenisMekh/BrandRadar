from __future__ import annotations

from pathlib import Path
from datetime import datetime

import torch
from torch.amp import autocast, GradScaler
from torch.utils.data import DataLoader, Dataset
from torch.optim import AdamW
from transformers import get_linear_schedule_with_warmup
from sklearn.metrics import f1_score, precision_score, recall_score

try:
    from tqdm import tqdm
except ImportError:
    tqdm = None

from model import CompanyRelevanceModel
from dev import (
    set_seed,
    load_config,
    create_or_load_model,
    prepare_data,
    evaluate,
)


class PreTokenizedDataset(Dataset):
    def __init__(
        self,
        texts: list[str],
        companies: list[str],
        labels: list[int],
        tokenizer,
        max_length: int = 256,
        keywords: list[str] | None = None,
    ):
        self.labels = labels
        self.keywords = keywords
        self.encodings = tokenizer(
            texts,
            companies if keywords is None else [f"{c}: {kw}" if kw else c for c, kw in zip(companies, keywords)],
            truncation=True,
            padding=True,
            max_length=max_length,
            return_tensors="pt",
        )

    def __len__(self):
        return len(self.labels)

    def __getitem__(self, idx):
        item = {k: v[idx] for k, v in self.encodings.items()}
        item["labels"] = torch.tensor(self.labels[idx], dtype=torch.long)
        return item


def collate_fn(batch: list[dict]) -> dict:
    keys = [k for k in batch[0] if k != "labels"]
    result = {k: torch.stack([b[k] for b in batch]) for k in keys}
    result["labels"] = torch.stack([b["labels"] for b in batch])
    return result


def compute_metrics(y_true, y_pred) -> dict:
    """Вычисляет все метрики из y_true / y_pred."""
    acc = sum(a == b for a, b in zip(y_true, y_pred)) / len(y_true)
    f1 = f1_score(y_true, y_pred, pos_label=1, zero_division=0)
    prec = precision_score(y_true, y_pred, pos_label=1, zero_division=0)
    rec = recall_score(y_true, y_pred, pos_label=1, zero_division=0)
    return {"accuracy": acc, "f1": f1, "precision": prec, "recall": rec}


def print_metrics_row(epoch: int, epochs: int, loss: float, metrics: dict,
                      lr_cls: float = None, lr_enc: float = None):
    """Компактный вывод метрик одной строкой."""
    parts = [f"Epoch {epoch}/{epochs}"]
    parts.append(f"loss={loss:.4f}")
    parts.append(f"acc={metrics['accuracy']:.4f}")
    parts.append(f"f1={metrics['f1']:.4f}")
    parts.append(f"prec={metrics['precision']:.4f}")
    parts.append(f"rec={metrics['recall']:.4f}")
    if lr_cls is not None:
        parts.append(f"lr_cls={lr_cls:.2e}")
    if lr_enc is not None:
        parts.append(f"lr_enc={lr_enc:.2e}")
    print("  " + "  |  ".join(parts))


def print_metrics_summary(history: list[dict]):
    """Таблица всех эпох в конце тренировки."""
    print(f"\n{'═'*90}")
    print("ИСТОРИЯ ТРЕНИРОВКИ")
    print(f"{'═'*90}")
    header = f"  {'Epoch':>5}  {'Loss':>8}  {'Acc':>8}  {'F1':>8}  {'Prec':>8}  {'Recall':>8}  {'Best':>4}"
    print(header)
    print(f"  {'─'*5}  {'─'*8}  {'─'*8}  {'─'*8}  {'─'*8}  {'─'*8}  {'─'*4}")

    best_f1 = max(h["f1"] for h in history)
    for h in history:
        mark = " " if h["f1"] == best_f1 else ""
        print(
            f"  {h['epoch']:>5}  {h['loss']:>8.4f}  {h['accuracy']:>8.4f}  "
            f"{h['f1']:>8.4f}  {h['precision']:>8.4f}  {h['recall']:>8.4f}  {mark}"
        )
    print(f"{'═'*90}\n")


def _train_epoch(
    rm: CompanyRelevanceModel,
    loader: DataLoader,
    optimizer: AdamW,
    scheduler,
    scaler: GradScaler,
    use_amp: bool,
    accum_steps: int,
    max_grad_norm: float = 1.0,
) -> float:
    rm.train_mode()
    total_loss = 0.0
    optimizer.zero_grad(set_to_none=True)

    iterator = tqdm(loader, desc="Training", unit="batch") if tqdm else loader

    for step, batch in enumerate(iterator):
        labels = batch.pop("labels")

        with autocast("cuda", enabled=use_amp):
            out = rm.forward_encoded(**batch, labels=labels)
            loss = out.loss / accum_steps

        scaler.scale(loss).backward()

        if (step + 1) % accum_steps == 0:
            scaler.unscale_(optimizer)
            params = [p for p in rm.parameters() if p.requires_grad]
            torch.nn.utils.clip_grad_norm_(params, max_grad_norm)
            scaler.step(optimizer)
            scaler.update()
            if scheduler is not None:
                scheduler.step()
            optimizer.zero_grad(set_to_none=True)

        total_loss += loss.item() * accum_steps

        if iterator is not None and hasattr(iterator, "set_postfix"):
            iterator.set_postfix(loss=loss.item() * accum_steps)

    if (step + 1) % accum_steps != 0:
        scaler.unscale_(optimizer)
        params = [p for p in rm.parameters() if p.requires_grad]
        torch.nn.utils.clip_grad_norm_(params, max_grad_norm)
        scaler.step(optimizer)
        scaler.update()
        if scheduler is not None:
            scheduler.step()
        optimizer.zero_grad(set_to_none=True)

    return total_loss / len(loader)


def train(cfg: dict):
    set_seed(cfg["train"]["seed"])
    tcfg = cfg["train"]

    print("=" * 60)
    print("Creating / loading model")
    print("=" * 60)
    rm = create_or_load_model(cfg)

    print("Preparing dataset …")
    train_df, test_df = prepare_data(cfg, rm, split=True)
    print(f"Train: {len(train_df)}  |  Test: {len(test_df)}")

    print("Evaluating before training …")
    acc_base, y_true_base, y_pred_base, _ = evaluate(rm, test_df, verbose=False)
    base_metrics = compute_metrics(y_true_base, y_pred_base)
    print(
        f"  Baseline: acc={base_metrics['accuracy']:.4f}  "
        f"f1={base_metrics['f1']:.4f}  "
        f"prec={base_metrics['precision']:.4f}  "
        f"rec={base_metrics['recall']:.4f}"
    )

    keywords_train = train_df["keywords"].tolist() if "keywords" in train_df.columns else None
    
    train_dataset = PreTokenizedDataset(
        texts=train_df["input"].tolist(),
        companies=train_df["company"].tolist(),
        labels=train_df["label_id"].tolist(),
        tokenizer=rm.tokenizer,
        max_length=rm.max_length,
        keywords=keywords_train,
    )

    num_workers = tcfg.get("num_workers", 4)
    train_loader = DataLoader(
        train_dataset,
        batch_size=tcfg["batch_size"],
        shuffle=True,
        collate_fn=collate_fn,
        num_workers=num_workers,
        pin_memory=rm.device.type == "cuda",
        persistent_workers=num_workers > 0,
        prefetch_factor=2 if num_workers > 0 else None,
    )

    use_amp = rm.device.type == "cuda"
    scaler = GradScaler("cuda", enabled=use_amp)
    accum_steps = tcfg.get("gradient_accumulation_steps", 1)
    max_grad_norm = tcfg.get("max_grad_norm", 1.0)

    if tcfg.get("torch_compile", False) and hasattr(torch, "compile"):
        print("Compiling encoder + classifier with torch.compile …")
        rm.encoder = torch.compile(rm.encoder, mode="reduce-overhead")
        rm.classifier = torch.compile(rm.classifier)

    history = []

    freeze_epochs = tcfg.get("freeze_encoder_epochs", 0)
    if freeze_epochs > 0:
        print(f"\n{'─'*60}")
        print(f"Phase 1: classifier only ({freeze_epochs} epochs, encoder frozen)")
        print(f"{'─'*60}")
        rm.freeze_encoder()
        head_lr = tcfg["learning_rate"] * tcfg.get("head_lr_multiplier", 10)
        opt_head = AdamW(rm.trainable_parameters(), lr=head_lr)

        for epoch in range(1, freeze_epochs + 1):
            avg_loss = _train_epoch(
                rm, train_loader, opt_head, None, scaler, use_amp, accum_steps, max_grad_norm
            )
            acc, y_true, y_pred, _ = evaluate(rm, test_df, verbose=False)
            metrics = compute_metrics(y_true, y_pred)
            metrics["epoch"] = f"W{epoch}"
            metrics["loss"] = avg_loss
            history.append(metrics)
            print_metrics_row(f"W{epoch}", freeze_epochs, avg_loss, metrics)

    epochs = tcfg["epochs"]
    print(f"\n{'─'*60}")
    print(f"Phase 2: full fine-tune ({epochs} epochs)")
    print(f"{'─'*60}")

    rm.unfreeze_encoder()

    encoder_lr = tcfg["learning_rate"] / tcfg.get("encoder_lr_divisor", 5)
    optimizer = AdamW(
        [
            {"params": rm.classifier.parameters(), "lr": tcfg["learning_rate"]},
            {"params": rm.encoder.parameters(), "lr": encoder_lr},
        ],
        weight_decay=tcfg.get("weight_decay", 0.01),
    )

    total_opt_steps = (len(train_loader) // accum_steps) * epochs
    warmup_steps = int(total_opt_steps * tcfg["warmup_ratio"])
    scheduler = get_linear_schedule_with_warmup(optimizer, warmup_steps, total_opt_steps)

    best_metric_name = tcfg.get("best_metric", "f1")
    best_metric_value = -1.0
    best_path = None

    for epoch in range(1, epochs + 1):
        avg_loss = _train_epoch(
            rm, train_loader, optimizer, scheduler, scaler, use_amp, accum_steps, max_grad_norm
        )

        lr_cls = scheduler.get_last_lr()[0]
        lr_enc = scheduler.get_last_lr()[-1]

        acc, y_true, y_pred, all_probs = evaluate(rm, test_df, verbose=True)
        metrics = compute_metrics(y_true, y_pred)
        metrics["epoch"] = epoch
        metrics["loss"] = avg_loss
        metrics["lr_cls"] = lr_cls
        metrics["lr_enc"] = lr_enc
        history.append(metrics)

        rm.train_mode()

        print_metrics_row(epoch, epochs, avg_loss, metrics, lr_cls, lr_enc)

        current_value = metrics[best_metric_name]
        if current_value > best_metric_value:
            best_metric_value = current_value
            version = datetime.now().strftime("%Y%m%d_%H%M%S")
            best_path = Path(cfg["output"]["models_dir"]) / f"v_{version}"
            rm.save(best_path)
            print(
                f"   New best {best_metric_name}={best_metric_value:.4f} -> {best_path}"
            )

    if best_path is None:
        version = datetime.now().strftime("%Y%m%d_%H%M%S")
        best_path = Path(cfg["output"]["models_dir"]) / f"v_{version}"
        rm.save(best_path)

    latest = Path(cfg["output"]["models_dir"]) / "latest"
    if latest.is_symlink() or latest.exists():
        latest.unlink()
    latest.symlink_to(best_path.resolve())

    print_metrics_summary(history)

    print(f"Best model ({best_metric_name}={best_metric_value:.4f}) -> {best_path}")
    print(f"Symlink latest -> {latest}")

    return rm, test_df


if __name__ == "__main__":
    cfg = load_config()
    rm, test_df = train(cfg)
    print("\n" + "=" * 60)
    print("ФИНАЛЬНАЯ ОЦЕНКА")
    print("=" * 60)
    acc, y_true, y_pred, _ = evaluate(rm, test_df, verbose=True)
    final = compute_metrics(y_true, y_pred)
    print(
        f"\nИтого: acc={final['accuracy']:.4f}  f1={final['f1']:.4f}  "
        f"prec={final['precision']:.4f}  rec={final['recall']:.4f}"
    )
