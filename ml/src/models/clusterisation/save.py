import logging
from pathlib import Path

import yaml

from model import TextClusterer

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)

CONFIG_PATH = Path("src/models/clusterisation/configs/dev_config.yml")


def main():
    with open(CONFIG_PATH, "r", encoding="utf-8") as f:
        cfg = yaml.safe_load(f)

    clusterer = TextClusterer.from_config(cfg, base_dir=CONFIG_PATH.parent)

    save_path = cfg["save"]["path"]
    clusterer.save(save_path)

    print(f"\n✅ Model saved to: {save_path}")
    print("Contents:")
    for f in sorted(Path(save_path).rglob("*")):
        if f.is_file():
            size_mb = f.stat().st_size / 1024 / 1024
            print(f"  {f.relative_to(save_path)}  ({size_mb:.1f} MB)")


if __name__ == "__main__":
    main()