from pathlib import Path
from datetime import datetime
from torch.utils.data import DataLoader
from torch.optim import AdamW
from transformers import get_linear_schedule_with_warmup
from tqdm import tqdm
from model import AspectSentimentModel

from dev import (
    set_seed,
    load_config,
    create_or_load_model,
    prepare_data,
    evaluate,
    AspectDataset,
    collate_fn,
)


def train(cfg: dict):
    set_seed(cfg["train"]["seed"])

    print("=" * 50)
    print("Creating / loading model")
    print("=" * 50)
    sm = create_or_load_model(cfg)

    print("Preparing dataset")
    train_df, test_df = prepare_data(cfg, sm, split=True)
    print(f"Train: {len(train_df)}, Test: {len(test_df)}")

    print("Evaluating before training")
    evaluate(sm, test_df, verbose=False)

    train_dataset = AspectDataset(
        train_df["input"].tolist(),
        train_df["company"].tolist(),
        train_df["label_id"].tolist(),
    )
    train_loader = DataLoader(
        train_dataset,
        batch_size=cfg["train"]["batch_size"],
        shuffle=True,
        collate_fn=collate_fn,
    )

    optimizer = AdamW(sm.parameters(), lr=cfg["train"]["learning_rate"])
    total_steps = len(train_loader) * cfg["train"]["epochs"]
    warmup_steps = int(total_steps * cfg["train"]["warmup_ratio"])
    scheduler = get_linear_schedule_with_warmup(optimizer, warmup_steps, total_steps)

    sm.train_mode()
    for epoch in range(cfg["train"]["epochs"]):
        total_loss = 0
        progress_bar = tqdm(train_loader, desc=f"Эпоха {epoch + 1}/{cfg['train']['epochs']}", unit="batch")
        for texts, aspects, labels in progress_bar:
            outputs = sm.forward(texts, aspects, labels)
            loss = outputs.loss
            loss.backward()
            optimizer.step()
            scheduler.step()
            optimizer.zero_grad()
            total_loss += loss.item()
            progress_bar.set_postfix({"loss": f"{loss.item():.4f}"})

        avg_loss = total_loss / len(train_loader)
        print(f"\nЭпоха {epoch + 1}/{cfg['train']['epochs']} завершена — средний loss: {avg_loss:.4f}")

        epoch_save_path = Path(cfg["output"]["models_dir"]) / f"epoch_{epoch + 1}"
        sm.save(epoch_save_path)
        print(f"Модель сохранена: {epoch_save_path}")

        evaluate(sm, test_df, verbose=True)
        sm.train_mode()

    version = datetime.now().strftime("%Y%m%d_%H%M%S")
    save_path = Path(cfg["output"]["models_dir"]) / f"v_{version}"
    sm.save(save_path)

    latest = Path(cfg["output"]["models_dir"]) / "latest"
    if latest.is_symlink() or latest.exists():
        latest.unlink()
    latest.symlink_to(save_path.resolve())
    print(f"Saved  {save_path}")

    return sm, test_df


if __name__ == "__main__":
    cfg = load_config()
    sm, test_df = train(cfg)
    evaluate(sm, test_df)
