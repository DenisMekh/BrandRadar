import asyncio
import aiohttp
import csv
import random
import os
import time
import logging
from dataclasses import dataclass
from typing import Optional

MAX_WORKERS = 40
TARGET_ROWS = 100_000
OUTPUT_FILE = "absa_dataset.csv"
CHECKPOINT_FILE = "absa_checkpoint.csv"
CHECKPOINT_EVERY = 500

# УКАЗАТЬ ДАННЫЕ
API_BASE_URL = ""
API_KEY = ""
MODEL_NAMES = [
    "",
]
API_TIMEOUT = 120
MAX_RETRIES = 5
RETRY_BASE_DELAY = 2

TEMPERATURE = 0.95
MAX_TOKENS = 2048

INDUSTRIES = {
    "банки": [
        "Сбербанк", "Тинькофф", "Альфа-Банк", "ВТБ", "Газпромбанк",
        "Райффайзен", "Росбанк", "Открытие", "Совкомбанк", "Промсвязьбанк",
        "Почта Банк", "Русский Стандарт", "Хоум Кредит", "МКБ", "Ренессанс"
    ],
    "телеком": [
        "МТС", "Билайн", "Мегафон", "Теле2", "Йота",
        "Ростелеком", "Дом.ру", "СберМобайл", "Тинькофф Мобайл", "Skylink"
    ],
    "маркетплейсы": [
        "Wildberries", "Ozon", "Яндекс Маркет", "СберМегаМаркет", "AliExpress",
        "Lamoda", "KazanExpress", "Avito", "Joom", "Deal.by"
    ],
    "доставка еды": [
        "Яндекс Еда", "Delivery Club", "Самокат", "Лавка", "ВкусВилл",
        "Сбермаркет", "Кухня на районе", "Перекрёсток Впрок", "iGooods", "Broniboy"
    ],
    "авиакомпании": [
        "Аэрофлот", "S7 Airlines", "Победа", "Уральские авиалинии", "Россия",
        "Nordwind", "Azur Air", "Red Wings", "Smartavia", "Якутия"
    ],
    "автомобили": [
        "LADA", "Haval", "Chery", "Geely", "EXEED",
        "Kia", "Hyundai", "Changan", "Omoda", "Москвич"
    ],
    "IT_компании": [
        "Яндекс", "VK", "Касперский", "1С", "Positive Technologies",
        "Сбер", "Ozon Tech", "Авито", "HeadHunter", "Wildberries Tech"
    ],
    "ритейл": [
        "Пятёрочка", "Магнит", "Лента", "Перекрёсток", "Ашан",
        "Метро", "Дикси", "ВкусВилл", "Азбука Вкуса", "Светофор"
    ],
    "фастфуд": [
        "Вкусно — и точка", "KFC", "Burger King", "Додо Пицца", "Subway",
        "Теремок", "Шоколадница", "Якитория", "Тануки", "Чайхона №1"
    ],
    "страхование": [
        "СОГАЗ", "Ингосстрах", "Росгосстрах", "АльфаСтрахование", "РЕСО-Гарантия",
        "Ренессанс Страхование", "Тинькофф Страхование", "Сбербанк Страхование",
        "ВСК", "Макс"
    ],
    "электроника": [
        "DNS", "М.Видео", "Эльдорадо", "Ситилинк", "re:Store",
        "Samsung", "Xiaomi", "Apple", "Huawei", "Honor"
    ],
    "такси": [
        "Яндекс Такси", "Uber", "Ситимобил", "Максим", "DiDi",
        "Gett", "inDriver", "Bolt", "Везёт", "Таксовичкоф"
    ],
    "фитнес": [
        "World Class", "X-Fit", "DDX Fitness", "Alex Fitness", "Spirit Fitness",
        "Фитнес Хаус", "Crocus Fitness", "WeGym", "Bright Fit", "С.С.С.Р."
    ],
    "образование": [
        "Skillbox", "Яндекс Практикум", "GeekBrains", "Нетология", "Skillfactory",
        "Otus", "Hexlet", "Stepik", "Учи.ру", "Skyeng"
    ],
    "медицина": [
        "Инвитро", "Гемотест", "Медси", "СМ-Клиника", "Европейский медицинский центр",
        "Мать и дитя", "Доктор рядом", "СберЗдоровье", "ПроДокторов", "DocDoc"
    ],
}

GENRES = [
    "новостная_статья",
    "пост_в_соцсетях",
    "комментарий_на_форуме",
    "отзыв_клиента",
    "сообщение_в_чате",
    "блог_пост",
    "сравнительный_обзор",
    "жалоба",
    "рекомендация_другу",
    "аналитическая_заметка",
    "тред_в_телеграме",
    "пост_на_пикабу",
    "ответ_в_комментариях",
    "обсуждение_на_vc_ru",
    "email_обращение",
    "тикток_описание",
    "сторис_текст",
    "твит",
]

SENTIMENT_COMBOS_2 = [
    ("positive", "positive"),
    ("positive", "negative"),
    ("positive", "neutral"),
    ("negative", "negative"),
    ("negative", "positive"),
    ("negative", "neutral"),
    ("neutral", "neutral"),
    ("neutral", "positive"),
    ("neutral", "negative"),
]

SENTIMENT_COMBOS_3 = [
    ("positive", "negative", "neutral"),
    ("positive", "positive", "negative"),
    ("negative", "negative", "positive"),
    ("positive", "neutral", "neutral"),
    ("negative", "neutral", "positive"),
    ("neutral", "neutral", "neutral"),
    ("positive", "positive", "positive"),
    ("negative", "negative", "negative"),
    ("positive", "negative", "negative"),
    ("negative", "positive", "positive"),
]

EVENTS_POSITIVE = [
    "запуск новой программы лояльности", "большой кешбэк", "удобное обновление приложения",
    "снижение цен", "быстрая доставка", "отличная служба поддержки",
    "выпуск нового продукта", "расширение сети", "удачная акция",
    "высокое качество обслуживания", "приятный бонус", "инновационная функция",
    "получение награды", "рост акций", "партнёрство с популярным брендом",
    "бесплатная подписка", "открытие нового филиала", "улучшение условий",
    "выгодная ипотечная ставка", "быстрое оформление", "щедрая реферальная программа",
    "стильный редизайн", "удобный интерфейс", "мгновенные переводы"
]

EVENTS_NEGATIVE = [
    "сбой в работе приложения", "утечка данных", "повышение цен",
    "плохая поддержка клиентов", "задержка доставки", "скрытые комиссии",
    "массовые увольнения", "скандал с руководством", "отказ в обслуживании",
    "низкое качество товара", "навязчивая реклама", "потеря посылки",
    "долгое ожидание ответа", "ошибка в начислениях", "блокировка аккаунтов",
    "мошенничество", "непрозрачные условия", "технические неполадки",
    "отмена рейса", "грубость персонала", "некорректное списание средств",
    "невозможность вернуть товар", "плохое качество связи", "падение сервера"
]

EVENTS_NEUTRAL = [
    "смена генерального директора", "ребрендинг", "обновление условий договора",
    "запуск новой рекламной кампании", "изменение графика работы",
    "переезд офиса", "выход квартального отчёта", "плановые технические работы",
    "участие в конференции", "обновление политики конфиденциальности",
    "смена логотипа", "выпуск пресс-релиза", "участие в выставке",
    "обновление тарифов", "запуск бета-версии", "назначение нового менеджера",
    "изменение юридического названия", "проведение внутреннего аудита"
]

AUTHOR_STYLES = [
    "обычный пользователь, пишет просто и по делу",
    "эмоциональный человек, использует восклицания и экспрессивную лексику",
    "аналитик, пишет сухо и с фактами",
    "молодой человек, использует сленг и мемы",
    "пожилой человек, пишет обстоятельно и подробно",
    "журналист, нейтральный тон, но с деталями",
    "IT-специалист, использует технические термины",
    "домохозяйка, делится бытовым опытом",
    "студент, пишет неформально",
    "бизнесмен, оценивает с точки зрения эффективности",
    "мама с ребёнком, фокус на удобстве",
    "путешественник, сравнивает с зарубежным опытом",
    "саркастичный блогер",
    "обеспокоенный гражданин",
    "лояльный клиент бренда",
]


def get_event_for_sentiment(sentiment: str) -> str:
    if sentiment == "positive":
        return random.choice(EVENTS_POSITIVE)
    elif sentiment == "negative":
        return random.choice(EVENTS_NEGATIVE)
    else:
        return random.choice(EVENTS_NEUTRAL)


def build_prompt(brands: list[str], sentiments: list[str], genre: str) -> str:
    """Build a diverse prompt for the LLM."""

    brand_descriptions = []
    for brand, sent in zip(brands, sentiments):
        event = get_event_for_sentiment(sent)
        sent_ru = {"positive": "позитивное", "negative": "негативное", "neutral": "нейтральное"}[sent]
        brand_descriptions.append(
            f'  - "{brand}": отношение {sent_ru}, связанное с событием: {event}'
        )
    brand_block = "\n".join(brand_descriptions)

    style = random.choice(AUTHOR_STYLES)

    length_hint = random.choice([
        "от 40 до 80 слов",
        "от 80 до 130 слов",
        "от 130 до 190 слов",
        "от 50 до 100 слов",
        "от 60 до 150 слов",
    ])

    genre_hints = {
        "новостная_статья": "Напиши как короткую новостную заметку с заголовком.",
        "пост_в_соцсетях": "Напиши как пост в социальной сети (VK, Facebook). Можно с эмодзи.",
        "комментарий_на_форуме": "Напиши как комментарий на форуме, в разговорном стиле.",
        "отзыв_клиента": "Напиши как отзыв клиента на сайте отзывов (irecommend, otzovik).",
        "сообщение_в_чате": "Напиши как сообщение в групповом чате (WhatsApp/Telegram). Неформально, с опечатками допустимы.",
        "блог_пост": "Напиши как запись в личном блоге.",
        "сравнительный_обзор": "Напиши как сравнительный мини-обзор, сопоставляя упомянутые компании.",
        "жалоба": "Напиши как жалобу или претензию.",
        "рекомендация_другу": "Напиши как если бы ты советовал(а) другу в личной переписке.",
        "аналитическая_заметка": "Напиши как короткую аналитическую заметку с цифрами/фактами.",
        "тред_в_телеграме": "Напиши как пост в Telegram-канале.",
        "пост_на_пикабу": "Напиши как пост на Пикабу, с юмором или историей.",
        "ответ_в_комментариях": "Напиши как ответ в комментариях к чьему-то посту.",
        "обсуждение_на_vc_ru": "Напиши как комментарий на vc.ru, деловой тон.",
        "email_обращение": "Напиши как email-обращение в компанию или обсуждение email'а.",
        "тикток_описание": "Напиши как описание к видео в TikTok, очень коротко.",
        "сторис_текст": "Напиши как текст для сторис — коротко, ёмко, с эмодзи.",
        "твит": "Напиши как твит или короткое сообщение до 280 символов. Может быть тред из 2-3 твитов.",
    }

    genre_instruction = genre_hints.get(genre, "Напиши в свободном стиле.")

    extra_instructions = random.choice([
        "",
        "\nУпомяни конкретную сумму денег или процент.",
        "\nДобавь сравнение с прошлым опытом.",
        "\nУпомяни конкретную дату или время.",
        "\nДобавь личную историю или анекдот.",
        "\nУпомяни конкретный город России.",
        "\nСсылайся на мнение знакомого или родственника.",
        "\nДобавь риторический вопрос.",
        "\nИспользуй ироничный тон.",
        "\nУпомяни конкурента одной из компаний.",
        "",
        "",
    ])

    prompt = f"""Ты генератор текстов для обучающего датасета анализа тональности (ABSA).

Задача: напиши ОДИН текст на русском языке в жанре "{genre}".
{genre_instruction}

Стиль автора: {style}

В тексте должны упоминаться следующие компании с заданным отношением:
{brand_block}

Требования:
1. Длина текста: {length_hint}.
2. Текст должен быть естественным и реалистичным, как будто его написал реальный человек.
3. Тональность к каждой компании должна быть понятна из контекста, но не слишком в лоб.
4. НЕ пиши явно слова "позитивно", "негативно", "нейтрально" — тональность должна считываться из смысла.
5. Компании должны упоминаться по названию.
6. Текст должен быть цельным и связным, а не набором отдельных предложений о каждой компании.{extra_instructions}

Ответь ТОЛЬКО текстом, без пояснений, кавычек и метаданных. Просто текст."""

    return prompt


@dataclass
class GeneratedSample:
    text: str
    brand: str
    sentiment: str


def pick_brands_and_sentiments() -> tuple[list[str], list[str]]:
    """Pick 2-3 brands from 1-2 industries with sentiment assignments."""
    num_brands = random.choices([2, 3], weights=[0.55, 0.45], k=1)[0]

    num_industries = random.choice([1, 2]) if num_brands >= 2 else 1
    chosen_industries = random.sample(list(INDUSTRIES.keys()), min(num_industries, len(INDUSTRIES)))

    brands = []
    for ind in chosen_industries:
        available = [b for b in INDUSTRIES[ind] if b not in brands]
        pick_count = min(len(available), max(1, num_brands - len(brands)))
        brands.extend(random.sample(available, pick_count))
        if len(brands) >= num_brands:
            break

    while len(brands) < num_brands:
        ind = random.choice(list(INDUSTRIES.keys()))
        available = [b for b in INDUSTRIES[ind] if b not in brands]
        if available:
            brands.append(random.choice(available))

    brands = brands[:num_brands]

    if num_brands == 2:
        sentiments = list(random.choice(SENTIMENT_COMBOS_2))
    else:
        sentiments = list(random.choice(SENTIMENT_COMBOS_3))

    return brands, sentiments


class LLMClient:
    def __init__(self, base_url: str, api_key: str, models: list[str]):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.models = models
        self.session: Optional[aiohttp.ClientSession] = None

    async def init_session(self):
        self.session = aiohttp.ClientSession(
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
            },
            timeout=aiohttp.ClientTimeout(total=API_TIMEOUT),
        )

    async def close(self):
        if self.session:
            await self.session.close()

    async def generate(self, prompt: str) -> Optional[str]:
        if not self.session:
            await self.init_session()

        model = random.choice(self.models)

        payload = {
            "model": model,
            "messages": [
                {"role": "user", "content": prompt}
            ],
            "temperature": TEMPERATURE,
            "max_tokens": MAX_TOKENS,
        }

        for attempt in range(MAX_RETRIES):
            try:
                async with self.session.post(
                    f"{self.base_url}/chat/completions",
                    json=payload,
                ) as resp:
                    if resp.status == 200:
                        data = await resp.json()
                        content = data["choices"][0]["message"]["content"]
                        return content.strip()
                    elif resp.status == 429:
                        delay = RETRY_BASE_DELAY * (2 ** attempt) + random.uniform(0, 1)
                        logging.warning(f"Rate limited, waiting {delay:.1f}s...")
                        await asyncio.sleep(delay)
                    else:
                        text = await resp.text()
                        logging.error(f"API error {resp.status}: {text[:200]}")
                        delay = RETRY_BASE_DELAY * (2 ** attempt)
                        await asyncio.sleep(delay)
            except asyncio.TimeoutError:
                logging.warning(f"Timeout on attempt {attempt + 1}")
                await asyncio.sleep(RETRY_BASE_DELAY * (2 ** attempt))
            except Exception as e:
                logging.error(f"Request error: {e}")
                await asyncio.sleep(RETRY_BASE_DELAY * (2 ** attempt))

        return None


def validate_text(text: str, brands: list[str]) -> bool:
    """Basic validation that text mentions all brands."""
    if not text or len(text) < 20:
        return False
    text_lower = text.lower()
    for brand in brands:
        brand_parts = brand.lower().split()
        if not any(part in text_lower for part in brand_parts):
            return False
    return True


async def worker(
    worker_id: int,
    client: LLMClient,
    result_queue: asyncio.Queue,
    rows_counter: dict,
    target_rows: int,
    lock: asyncio.Lock,
):
    """Worker that generates samples and puts them into result_queue."""
    while True:
        async with lock:
            if rows_counter["count"] >= target_rows:
                return

        brands, sentiments = pick_brands_and_sentiments()
        genre = random.choice(GENRES)
        prompt = build_prompt(brands, sentiments, genre)

        text = await client.generate(prompt)

        if text is None:
            logging.warning(f"[Worker {worker_id}] Failed to generate, retrying...")
            continue

        text = text.strip().strip('"').strip("«»")


        samples = []
        for brand, sentiment in zip(brands, sentiments):
            samples.append(GeneratedSample(text=text, brand=brand, sentiment=sentiment))

        async with lock:
            remaining = target_rows - rows_counter["count"]
            if remaining <= 0:
                return
            to_add = samples[:remaining]
            rows_counter["count"] += len(to_add)
            for s in to_add:
                await result_queue.put(s)

            if rows_counter["count"] % 100 < len(to_add):
                logging.info(
                    f"Progress: {rows_counter['count']}/{target_rows} rows "
                    f"({100*rows_counter['count']/target_rows:.1f}%)"
                )


async def csv_writer(
    result_queue: asyncio.Queue,
    output_file: str,
    target_rows: int,
    rows_written: dict,
):
    """Async writer that reads from queue and writes to CSV."""

    file_exists = os.path.exists(output_file) and os.path.getsize(output_file) > 0

    mode = "a" if file_exists else "w"
    f = open(output_file, mode, newline="", encoding="utf-8")
    writer = csv.writer(f, quoting=csv.QUOTE_ALL)

    if not file_exists:
        writer.writerow(["text", "brand", "sentiment"])
        f.flush()

    batch = []
    batch_count = 0

    try:
        while rows_written["count"] < target_rows:
            try:
                sample = await asyncio.wait_for(result_queue.get(), timeout=30)
            except asyncio.TimeoutError:
                continue

            clean_text = sample.text.replace("\n", " ").replace("\r", " ")
            clean_text = " ".join(clean_text.split())

            writer.writerow([clean_text, sample.brand, sample.sentiment])
            rows_written["count"] += 1
            batch.append(1)

            if len(batch) >= 50:
                f.flush()
                batch.clear()
                batch_count += 1

                if batch_count % CHECKPOINT_EVERY == 0:
                    logging.info(f"Checkpoint: {rows_written['count']} rows written to {output_file}")
    finally:
        f.flush()
        f.close()
        logging.info(f"CSV writer finished. Total rows written: {rows_written['count']}")


async def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s [%(levelname)s] %(message)s",
        datefmt="%H:%M:%S",
    )

    logging.info(f"Starting ABSA dataset generation")
    logging.info(f"Target: {TARGET_ROWS} rows | Workers: {MAX_WORKERS}")
    logging.info(f"Output: {OUTPUT_FILE}")
    logging.info(f"API: {API_BASE_URL} | Models: {', '.join(MODEL_NAMES)}")

    existing_rows = 0
    if os.path.exists(OUTPUT_FILE):
        with open(OUTPUT_FILE, "r", encoding="utf-8") as f:
            reader = csv.reader(f)
            existing_rows = max(0, sum(1 for _ in reader) - 1)
        logging.info(f"Found {existing_rows} existing rows in {OUTPUT_FILE}")

    remaining = TARGET_ROWS - existing_rows
    if remaining <= 0:
        logging.info("Target already reached!")
        return

    logging.info(f"Need to generate {remaining} more rows")

    result_queue = asyncio.Queue(maxsize=MAX_WORKERS * 10)
    rows_counter = {"count": 0}
    rows_written = {"count": existing_rows}
    lock = asyncio.Lock()

    clients = []
    for _ in range(MAX_WORKERS):
        client = LLMClient(API_BASE_URL, API_KEY, MODEL_NAMES)
        await client.init_session()
        clients.append(client)

    worker_tasks = []
    for i in range(MAX_WORKERS):
        task = asyncio.create_task(
            worker(i, clients[i], result_queue, rows_counter, remaining, lock)
        )
        worker_tasks.append(task)

    writer_task = asyncio.create_task(
        csv_writer(result_queue, OUTPUT_FILE, remaining + existing_rows, rows_written)
    )

    start_time = time.time()

    await asyncio.gather(*worker_tasks)

    while not result_queue.empty():
        await asyncio.sleep(0.5)

    await asyncio.sleep(2)
    writer_task.cancel()
    try:
        await writer_task
    except asyncio.CancelledError:
        pass

    for client in clients:
        await client.close()

    elapsed = time.time() - start_time
    total_written = rows_written["count"]
    rate = (total_written - existing_rows) / elapsed if elapsed > 0 else 0

    logging.info(f"=" * 60)
    logging.info(f"Generation complete!")
    logging.info(f"Total rows in file: {total_written}")
    logging.info(f"New rows generated: {total_written - existing_rows}")
    logging.info(f"Time elapsed: {elapsed:.1f}s ({elapsed/60:.1f}min)")
    logging.info(f"Rate: {rate:.1f} rows/sec")
    logging.info(f"Output: {OUTPUT_FILE}")


if __name__ == "__main__":
    asyncio.run(main())
