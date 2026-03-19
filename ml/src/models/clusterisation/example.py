import logging
from model import TextClusterer

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-7s | %(name)s | %(message)s",
    datefmt="%H:%M:%S",
)


def main():
    clusterer = TextClusterer(
        model_name="ai-forever/sbert_large_nlu_ru",
        max_length=256,
        normalize_embeddings=True,
        batch_size=32,
        device=None,
        show_progress_bar=False,

        stopwords=[
            "т-банк", "т банк", "тбанк", "tbank", "t-bank", "тинькофф", "tinkoff",
            "банк", "банка", "банке",
        ],
        remove_tokens=[
            "Москве-Сити", "Петербурге", "Казани",
            "Т-Банк", "т-банк", "т банк", "t-bank", "t bank", "tbank", "тинькофф", "tinkoff",
        ],
        preprocess_lowercase=True,
        preprocess_remove_punctuation=True,
        preprocess_remove_stopwords=True,
        preprocess_min_length=3,

        use_umap=True,
        umap_n_neighbors=12,
        umap_n_components=5,
        umap_min_dist=0.0,
        umap_metric="cosine",

        topic_keywords_top_n=6,
    )

    messages = [
        "Т-Банк приложение вылетает при входе — что делать?",
        "Не могу зайти в приложение Т-Банка после обновления",
        "Приложение Т-Банка зависает на экране загрузки",
        "Т-Банк не отправляет push-уведомления о транзакциях",
        "Не приходит код подтверждения в приложении Т-Банка",
        "Не работает вход по биометрии в приложении Т-Банка",

        "Т-Банк объявил новые категории с повышенным кэшбэком на март",
        "Как выбрать категории кэшбэка в приложении Т-Банка?",
        "Т-Банк не начислил кэшбэк за покупку в супермаркете",
        "Почему Т-Банк не засчитал покупку в категории с повышенным кэшбэком?",
        "Максимальный лимит кэшбэка в Т-Банке: сколько можно вернуть?",

        "Т-Банк открыл новый офис в Москве-Сити — стоит заходить?",
        "Где находится новый флагманский офис Т-Банка в Петербурге?",
        "Как записаться на консультацию в новый офис Т-Банка?",
        "Открылся новый офис Т-Банка в Казани — какие услуги доступны?",

        "Т-Банк: как открыть ИИС и получить налоговый вычет?",
        "Т-Банк Инвестиции: какие комиссии за сделки с акциями?",
        "Как купить облигации федерального займа через Т-Банк?",
        "Т-Банк не выводит дивиденды на брокерский счёт — куда писать?",

        "Т-Банк одобрил кредитную карту с лимитом 300 000 ₽ — какие условия?",
        "Т-Банк: можно ли увеличить лимит по кредитной карте онлайн?",
        "Т-Банк не списывает минимальный платёж по кредитке — что делать?",
        "Просрочил платёж по кредиту в Т-Банке на 1 день — будут ли штрафы?",
    ]

    result = clusterer.cluster(
        texts=messages,
        min_cluster_size=3,
        min_samples=2,
        cluster_selection_epsilon=0.0,
        return_embeddings=False,
    )

    print("\n=== RESULTS ===")
    print(f"Total messages: {len(messages)}")
    print(f"Clusters: {len(result.clusters)} | Noise: {len(result.noise_messages)}")
    if result.silhouette is not None:
        print(f"Silhouette: {result.silhouette:.3f}")

    print("\n=== CLUSTERS ===")
    for c in sorted(result.clusters, key=lambda x: x.size, reverse=True):
        name = " · ".join(c.topic_keywords[:3]) if c.topic_keywords else "Без темы"
        print(f"\n[Cluster {c.cluster_id}] size={c.size} | {name}")
        if c.topic_keywords:
            print(f"  keywords: {', '.join(c.topic_keywords)}")
        for i, m in enumerate(c.messages, 1):
            print(f"  {i}. {m}")

    if result.noise_messages:
        print("\n=== NOISE ===")
        for i, m in enumerate(result.noise_messages, 1):
            print(f"  {i}. {m}")


if __name__ == "__main__":
    main()