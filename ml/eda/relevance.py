import asyncio
import aiohttp
import csv
import json
import random
import os
import time
import logging
import re
from dataclasses import dataclass
from typing import Optional

from aiohttp import payload_type
from tqdm import tqdm

MAX_WORKERS = 20
TARGET_ROWS = 50_000
OUTPUT_FILE = "relevance_v3.csv"
CHECKPOINT_EVERY = 200

# УКАЗАТЬ ДАННЫЕ
API_BASE_URL = ""
API_KEY = ""
MODEL_NAME = ""

API_TIMEOUT = 120
MAX_RETRIES = 5
RETRY_BASE_DELAY = 2

TEMPERATURE = 0.95
MAX_TOKENS = 2048

COMPANIES_PER_TEXT = 5


def sanitize_text(text: str) -> str:
    text = text.replace("‑", "-").replace("—", "-").replace("–", "-")
    text = re.sub(r'[\u200b\u200c\u200d\ufeff]', '', text)
    text = text.replace("[", "").replace("]", "")
    return text


@dataclass
class Company:
    name: str
    keywords: list[str]
    industry: str
    description: str


COMPANIES = [
    Company(
        name="Т-Банк",
        keywords=[
            "Тинькофф", "Tinkoff", "Т-Банк", "Т‑Банк", "TBank",
            "Т-Страхование", "Т-Инвестиции", "Т-Бизнес",
            "Оливер Хьюз", "тинькова", "Тинькофф Банк",
            "Тинькофф Мобайл", "Тинькофф Платинум",
            "тинёк", "жёлтый банк",
        ],
        industry="банки",
        description="онлайн-банк, бывший Тинькофф, CEO Станислав Близнюк",
    ),
    Company(
        name="Сбер",
        keywords=[
            "Сбербанк", "Sber", "Сбер", "SberBank",
            "СберПрайм", "СберЗдоровье", "СберМаркет",
            "СберМобайл", "СберИнвестиции",
            "Греф", "Герман Греф",
            "GigaChat", "Салют", "сберкот",
        ],
        industry="банки",
        description="крупнейший банк России, CEO Герман Греф, экосистема с GigaChat",
    ),
    Company(
        name="Яндекс",
        keywords=[
            "Яндекс", "Yandex",
            "Яндекс Такси", "Яндекс Еда", "Яндекс Маркет",
            "Яндекс Музыка", "Яндекс Плюс", "Яндекс Практикум",
            "Яндекс Браузер", "Яндекс Карты", "Кудрин",
            "YandexGPT", "Алиса",
        ],
        industry="IT",
        description="крупнейшая IT-компания России, поисковик, голосовой помощник Алиса",
    ),
    Company(
        name="Wildberries",
        keywords=[
            "Wildberries", "Вайлдберриз", "ВБ", "WB",
            "Бакальчук", "Татьяна Бакальчук",
            "вайлдбериз", "вилдберис", "Russ",
        ],
        industry="маркетплейсы",
        description="крупнейший маркетплейс России, основатель Татьяна Бакальчук",
    ),
    Company(
        name="Ozon",
        keywords=[
            "Ozon", "Озон", "OZON",
            "Ozon Fresh", "Ozon Банк", "Ozon Seller",
            "Шульгин", "Сергей Белл",
            "ozon.ru",
        ],
        industry="маркетплейсы",
        description="крупный маркетплейс и финтех",
    ),
    Company(
        name="МТС",
        keywords=[
            "МТС", "MTS",
            "МТС Банк", "МТС Музыка", "МТС Premium", "МТС Линк",
            "Евтушенков", "Вячеслав Николаев",
            "мобильные телесистемы",
        ],
        industry="телеком",
        description="телеком-оператор и экосистема, президент Вячеслав Николаев",
    ),
    Company(
        name="Мегафон",
        keywords=[
            "Мегафон", "MegaFon", "Megafon",
            "Мегафон ТВ", "Хачатурянц",
            "Усманов", "Алишер Усманов",
        ],
        industry="телеком",
        description="телеком-оператор, связан с USM Алишера Усманова",
    ),
    Company(
        name="Аэрофлот",
        keywords=[
            "Аэрофлот", "Aeroflot",
            "Аэрофлот Бонус", "Победа", "Россия авиа",
            "Александровский", "Сергей Александровский",
            "Шереметьево",
        ],
        industry="авиакомпании",
        description="крупнейшая авиакомпания, дочерние Победа и Россия, хаб Шереметьево",
    ),
    Company(
        name="VK",
        keywords=[
            "VK", "ВКонтакте", "ВК", "vk.com",
            "VK Видео", "VK Музыка", "VK Play", "VK Cloud",
            "Одноклассники", "Кириенко", "Владимир Кириенко",
            "Mail.ru",
        ],
        industry="IT",
        description="IT-компания, соцсети ВКонтакте и Одноклассники, CEO Владимир Кириенко",
    ),
    Company(
        name="Газпром",
        keywords=[
            "Газпром", "Gazprom",
            "Газпром нефть", "Газпромбанк", "Газпром-Медиа",
            "Миллер", "Алексей Миллер",
            "национальное достояние",
        ],
        industry="энергетика",
        description="крупнейшая газовая компания, CEO Алексей Миллер",
    ),
    Company(
        name="Kaspersky",
        keywords=[
            "Касперский", "Kaspersky",
            "Лаборатория Касперского", "KasperskyOS",
            "Евгений Касперский",
            "антивирус Касперского",
        ],
        industry="кибербезопасность",
        description="кибербезопасность, основатель Евгений Касперский",
    ),
    Company(
        name="Додо Пицца",
        keywords=[
            "Додо Пицца", "Dodo Pizza", "Додо",
            "Dodo Brands", "Овчинников", "Фёдор Овчинников",
            "Дринкит",
        ],
        industry="общепит",
        description="сеть пиццерий, основатель Фёдор Овчинников",
    ),
    Company(
        name="Роснефть",
        keywords=[
            "Роснефть", "Rosneft",
            "Сечин", "Игорь Сечин",
            "Башнефть",
        ],
        industry="энергетика",
        description="нефтяная компания, CEO Игорь Сечин",
    ),
    Company(
        name="Альфа-Банк",
        keywords=[
            "Альфа-Банк", "Alfa-Bank", "Альфа банк",
            "Альфа-Инвестиции", "Альфа-Страхование",
            "Фридман", "Михаил Фридман",
            "Авен", "Пётр Авен", "Альфа-Групп",
        ],
        industry="банки",
        description="крупный частный банк, основатели Михаил Фридман и Пётр Авен",
    ),
    Company(
        name="Positive Technologies",
        keywords=[
            "Positive Technologies", "PT",
            "MaxPatrol", "PT NGFW", "PT Sandbox",
            "Лукацкий", "Денис Баранов",
            "POSI", "Positive Hack Days", "PHDays",
        ],
        industry="кибербезопасность",
        description="кибербезопасность, CEO Денис Баранов, конференция PHDays",
    ),
]

COMPANY_INDEX = {c.name: c for c in COMPANIES}
ALL_NAMES = [c.name for c in COMPANIES]

GENRES = [
    "новостная_статья", "пост_в_соцсетях", "комментарий_на_форуме",
    "отзыв_клиента", "сообщение_в_чате", "блог_пост",
    "аналитическая_заметка", "тред_в_телеграме", "пост_на_пикабу",
    "обсуждение_на_vc_ru", "твит", "email", "пресс_релиз",
    "жалоба", "рекомендация", "сравнительный_обзор",
]

GENRE_HINTS = {
    "новостная_статья": "Напиши как короткую новостную заметку.",
    "пост_в_соцсетях": "Напиши как пост в VK/Telegram, можно с эмодзи.",
    "комментарий_на_форуме": "Напиши как комментарий на форуме.",
    "отзыв_клиента": "Напиши как отзыв на сайте отзывов.",
    "сообщение_в_чате": "Напиши как сообщение в чате, неформально.",
    "блог_пост": "Напиши как запись в блоге.",
    "аналитическая_заметка": "Напиши как аналитику с фактами и цифрами.",
    "тред_в_телеграме": "Напиши как пост в Telegram-канале.",
    "пост_на_пикабу": "Напиши как пост на Пикабу.",
    "обсуждение_на_vc_ru": "Напиши как комментарий на vc.ru.",
    "твит": "Напиши коротко, как твит или тред из 2-3 твитов.",
    "email": "Напиши как email-сообщение.",
    "пресс_релиз": "Напиши в стиле пресс-релиза.",
    "жалоба": "Напиши как жалобу.",
    "рекомендация": "Напиши как совет другу.",
    "сравнительный_обзор": "Напиши как сравнительный обзор.",
}

STYLES = [
    "обычный пользователь", "эмоциональный с восклицаниями",
    "сухой аналитик", "молодой со сленгом", "журналист",
    "IT-специалист", "домохозяйка", "студент",
    "бизнесмен", "саркастичный блогер", "пенсионер", "инвестор",
    "мама с ребёнком", "путешественник", "фрилансер",
]

OFF_TOPIC_THEMES = [
    "прогноз погоды на неделю в Москве",
    "рецепт борща по-домашнему",
    "обзор нового фильма Marvel",
    "результаты матча Спартак — ЦСКА",
    "советы по выращиванию помидоров на даче",
    "обзор новой модели iPhone",
    "путешествие по Грузии: маршрут на 7 дней",
    "как выбрать корм для кошки",
    "новости науки: открытие экзопланеты",
    "тренды моды осень 2024",
    "ремонт в квартире: выбор ламината",
    "обзор компьютерной игры Baldur's Gate 3",
    "как начать бегать по утрам",
    "история Санкт-Петербурга",
    "философия стоицизма для начинающих",
    "как приготовить тирамису дома",
    "отзыв о книге «Мастер и Маргарита»",
    "поездка на Байкал летом",
    "уход за комнатными растениями зимой",
    "сравнение беспроводных наушников",
    "как научить ребёнка читать",
    "обзор электросамоката Ninebot",
    "рецепт шарлотки с яблоками",
    "новости чемпионата мира по хоккею",
    "как медитировать: гайд для новичков",
    "топ-10 сериалов Netflix 2024",
    "выбор палатки для похода",
    "как оформить загранпаспорт",
    "обзор робота-пылесоса Roborock",
    "советы начинающим фотографам",
]

EVENTS_POSITIVE = [
    "запуск кешбэка 10%", "обновление приложения с удобным дизайном",
    "снижение комиссий", "открытие нового сервиса", "рост акций на 15%",
    "партнёрство с крупным брендом", "запуск бесплатной подписки",
    "победа в рейтинге качества", "рекордная выручка за квартал",
    "запуск AI-помощника", "расширение в новые регионы",
]

EVENTS_NEGATIVE = [
    "масштабный сбой приложения", "утечка данных клиентов",
    "повышение тарифов на 30%", "скандал с руководством",
    "массовые жалобы на поддержку", "блокировка аккаунтов",
    "задержки доставки на неделю", "скрытые комиссии",
    "падение акций на 20%", "увольнение сотрудников",
    "судебный иск от клиентов",
]

EVENTS_NEUTRAL = [
    "смена генерального директора", "ребрендинг логотипа",
    "квартальный отчёт без сюрпризов", "плановые техработы",
    "участие в конференции", "обновление пользовательского соглашения",
    "переезд головного офиса", "выход на IPO",
    "запуск бета-тестирования", "назначение нового CTO",
]

EXTRAS = [
    "", "", "",
    "\nУпомяни конкретную дату или месяц.",
    "\nДобавь личную историю.",
    "\nУпомяни российский город.",
    "\nДобавь конкретные цифры.",
    "\nИспользуй ироничный тон.",
    "\nДобавь сравнение с зарубежным аналогом.",
    "\nУпомяни мнение знакомого.",
    "\nДобавь риторический вопрос.",
]


@dataclass
class TextScenario:
    """
    Describes what text to generate and which companies to evaluate.
    """
    mentioned_companies: list[Company]
    check_companies: list[Company]
    relevance_map: dict
    text_type: str
    off_topic_theme: str
    mention_style: dict


def pick_scenario() -> TextScenario:
    """
    Three types of scenarios:
    1. corporate_multi (70%): text about 2-3 companies, check 5 (some relevant, some not)
    2. corporate_single (15%): text about 1 company, check 5
    3. off_topic (15%): text about weather/cooking/etc, check 5 (all irrelevant)
    """
    roll = random.random()

    if roll < 0.15:
        check = random.sample(COMPANIES, COMPANIES_PER_TEXT)
        return TextScenario(
            mentioned_companies=[],
            check_companies=check,
            relevance_map={c.name: False for c in check},
            text_type="off_topic",
            off_topic_theme=random.choice(OFF_TOPIC_THEMES),
            mention_style={},
        )

    elif roll < 0.30:
        mentioned = random.sample(COMPANIES, 1)
        remaining = [c for c in COMPANIES if c.name != mentioned[0].name]
        extras = random.sample(remaining, COMPANIES_PER_TEXT - 1)
        check = mentioned + extras
        random.shuffle(check)

        mention_styles = ["direct", "keyword", "leader", "product", "slang"]
        style_map = {mentioned[0].name: random.choice(mention_styles)}

        return TextScenario(
            mentioned_companies=mentioned,
            check_companies=check,
            relevance_map={c.name: (c.name == mentioned[0].name) for c in check},
            text_type="corporate_single",
            off_topic_theme="",
            mention_style=style_map,
        )

    else:
        num_mentioned = random.choices([2, 3], weights=[0.6, 0.4], k=1)[0]

        if random.random() < 0.3:
            industry = random.choice(list(set(c.industry for c in COMPANIES)))
            same_ind = [c for c in COMPANIES if c.industry == industry]
            if len(same_ind) >= num_mentioned:
                mentioned = random.sample(same_ind, num_mentioned)
            else:
                mentioned = random.sample(COMPANIES, num_mentioned)
        else:
            mentioned = random.sample(COMPANIES, num_mentioned)

        mentioned_names = {c.name for c in mentioned}
        remaining = [c for c in COMPANIES if c.name not in mentioned_names]

        num_extras = COMPANIES_PER_TEXT - num_mentioned

        same_ind_extras = [c for c in remaining
                          if c.industry in {m.industry for m in mentioned}]
        diff_ind_extras = [c for c in remaining if c not in same_ind_extras]

        extras = []
        si_count = min(len(same_ind_extras), random.randint(1, 2), num_extras)
        extras.extend(random.sample(same_ind_extras, si_count))
        left = num_extras - len(extras)
        if left > 0:
            pool = [c for c in remaining if c not in extras]
            extras.extend(random.sample(pool, min(left, len(pool))))

        check = mentioned + extras
        random.shuffle(check)

        mention_styles_options = ["direct", "keyword", "leader", "product", "slang"]
        style_map = {c.name: random.choice(mention_styles_options) for c in mentioned}

        return TextScenario(
            mentioned_companies=mentioned,
            check_companies=check,
            relevance_map={c.name: (c.name in mentioned_names) for c in check},
            text_type="corporate_multi",
            off_topic_theme="",
            mention_style=style_map,
        )


def _mention_instruction(company: Company, style: str) -> str:
    """How to mention a specific company in the text."""
    if style == "direct":
        return f'Упомяни "{company.name}" напрямую по названию.'
    elif style == "keyword":
        alt = [k for k in company.keywords if k.lower() != company.name.lower()]
        kw = random.choice(alt) if alt else company.name
        return f'Упомяни через ключевое слово "{kw}" (это {company.name}). Основное название можно не писать.'
    elif style == "leader":
        leaders = [k for k in company.keywords if len(k.split()) <= 2 and k[0].isupper() and k != company.name]
        leader = random.choice(leaders) if leaders else company.keywords[0]
        return f'Упомяни руководителя/основателя {leader} в контексте {company.name}.'
    elif style == "product":
        products = [k for k in company.keywords if len(k) > 4 and k != company.name]
        prod = random.choice(products) if products else company.name
        return f'Упомяни продукт/сервис "{prod}" компании {company.name}.'
    elif style == "slang":
        slang = [k for k in company.keywords if k.islower() and k != company.name.lower()]
        s = random.choice(slang) if slang else company.name
        return f'Упомяни разговорно как "{s}" (это {company.name}).'
    return f'Упомяни "{company.name}".'


def build_prompt(scenario: TextScenario) -> str:
    genre = random.choice(GENRES)
    genre_hint = GENRE_HINTS.get(genre, "")
    style = random.choice(STYLES)
    extra = random.choice(EXTRAS)

    length_hint = random.choice([
        "от 40 до 80 слов", "от 60 до 120 слов",
        "от 80 до 150 слов", "от 100 до 190 слов",
    ])

    if scenario.text_type == "off_topic":
        prompt = f"""Ты создаёшь тексты для обучения нейросети.

Сгенерируй ОДИН текст на русском языке. Тема: "{scenario.off_topic_theme}".

Жанр: {genre}. {genre_hint}
Стиль: {style}.
Объём: {length_hint}.

КРИТИЧЕСКИ ВАЖНО: пиши СТРОГО по заданной теме.
ЗАПРЕЩЕНО упоминать коммерческие организации: банки, IT-компании, маркетплейсы, телеком-операторы и другие бренды.
Текст должен быть максимально естественным и жизненным, без намёка на рекламу.{extra}

Выведи ТОЛЬКО готовый текст, без кавычек и дополнительных комментариев."""
        return prompt

    mentioned = scenario.mentioned_companies
    not_mentioned = [c for c in scenario.check_companies
                     if not scenario.relevance_map[c.name]]

    mention_lines = []
    for c in mentioned:
        ms = scenario.mention_style.get(c.name, "direct")
        event_pool = random.choice([EVENTS_POSITIVE, EVENTS_NEGATIVE, EVENTS_NEUTRAL])
        event = random.choice(event_pool)
        instruction = _mention_instruction(c, ms)
        mention_lines.append(f"  • {c.name} ({c.description}): {instruction} Событие: {event}.")

    mention_block = "\n".join(mention_lines)

    avoid_names = [c.name for c in not_mentioned]
    avoid_keywords = []
    for c in not_mentioned:
        avoid_keywords.extend(c.keywords[:3])
    avoid_str = ", ".join(avoid_names)
    avoid_kw_str = ", ".join(random.sample(avoid_keywords, min(8, len(avoid_keywords))))

    prompt = f"""Твоя задача — создавать тексты для обучающего датасета.

Напиши ОДИН развёрнутый текст на русском языке.

Жанр: {genre}. {genre_hint}
Манера письма: {style}.
Размер: {length_hint}.

Обязательно упомяни в тексте следующие компании:
{mention_block}

КАТЕГОРИЧЕСКИ НЕЛЬЗЯ упоминать эти компании и связанные с ними слова:
  Запрещённые компании: {avoid_str}
  Запрещённые ключевые слова: {avoid_kw_str}

Основные требования:
1. Пиши максимально естественно и живо.
2. Исключи слова "релевантный", "нерелевантный", "датасет" и им подобные.
3. Текст должен быть единым повествованием, без разрывов.
4. Все указанные компании должны быть вплетены в сюжет органично, без натяжек.{extra}

Пришли ТОЛЬКО финальный текст, без каких-либо пояснений, кавычек или мета-комментариев."""

    return prompt


@dataclass
class GeneratedRow:
    text: str
    company: str
    keywords: str
    relevant: int


class LLMClient:
    def __init__(self, base_url: str, api_key: str, model: str):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.model = model
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

        prompt = sanitize_text(prompt)

        payload = {
            "model": self.model,
            "messages": [{"role": "user", "content": prompt}],
            "temperature": TEMPERATURE,
            "max_tokens": MAX_TOKENS,
        }

        for attempt in range(MAX_RETRIES):
            try:
                async with self.session.post(
                    f"{self.base_url}/chat/completions", json=payload,
                ) as resp:
                    if resp.status == 200:
                        data = await resp.json()
                        return data["choices"][0]["message"]["content"].strip()
                    elif resp.status == 429:
                        delay = RETRY_BASE_DELAY * (2 ** attempt) + random.uniform(0, 1)
                        logging.warning(f"Rate limited, wait {delay:.1f}s")
                        await asyncio.sleep(delay)
                    else:
                        text = await resp.text()
                        logging.error(f"API {resp.status}: {text[:200]}")
                        await asyncio.sleep(RETRY_BASE_DELAY * (2 ** attempt))
            except asyncio.TimeoutError:
                await asyncio.sleep(RETRY_BASE_DELAY * (2 ** attempt))
            except Exception as e:
                logging.error(f"Error: {e}")
                await asyncio.sleep(RETRY_BASE_DELAY * (2 ** attempt))
        return None


def validate_text(text: str, scenario: TextScenario) -> bool:
    """Validate that text matches the scenario expectations."""
    if not text or len(text.split()) < 8:
        return False

    text_lower = text.lower()

    for c in scenario.mentioned_companies:
        found = False
        for kw in c.keywords:
            if kw.lower() in text_lower:
                found = True
                break
        if not found:
            style = scenario.mention_style.get(c.name, "direct")
            if style not in ("slang", "product", "leader"):
                return False

    for c in scenario.check_companies:
        if scenario.relevance_map[c.name]:
            continue
        if c.name.lower() in text_lower:
            return False
        for kw in c.keywords[:2]:
            if kw.lower() in text_lower:
                return False

    return True


progress_bar: Optional[tqdm] = None

async def worker(
    worker_id: int,
    client: LLMClient,
    result_queue: asyncio.Queue,
    rows_counter: dict,
    target_rows: int,
    lock: asyncio.Lock,
    pbar: tqdm,
):
    while True:
        async with lock:
            if rows_counter["count"] >= target_rows:
                return

        scenario = pick_scenario()
        prompt = build_prompt(scenario)
        text = await client.generate(prompt)

        if text is None:
            logging.warning(f"[W{worker_id}] Generation failed")
            continue

        text = text.strip().strip('"').strip("«»")

        if not validate_text(text, scenario):
            logging.debug(f"[W{worker_id}] Validation failed, skipping")
            continue

        clean_text = " ".join(text.replace("\n", " ").replace("\r", " ").split())

        rows = []
        for c in scenario.check_companies:
            rows.append(GeneratedRow(
                text=clean_text,
                company=c.name,
                keywords="; ".join(c.keywords),
                relevant=1 if scenario.relevance_map[c.name] else 0,
            ))

        async with lock:
            remaining = target_rows - rows_counter["count"]
            if remaining <= 0:
                return
            to_add = rows[:remaining]
            rows_counter["count"] += len(to_add)
            for r in to_add:
                await result_queue.put(r)
                pbar.update(1)


async def csv_writer(
    result_queue: asyncio.Queue,
    output_file: str,
    target_rows: int,
    rows_written: dict,
):
    file_exists = os.path.exists(output_file) and os.path.getsize(output_file) > 0
    mode = "a" if file_exists else "w"
    f = open(output_file, mode, newline="", encoding="utf-8")
    writer = csv.writer(f, quoting=csv.QUOTE_ALL)

    if not file_exists:
        writer.writerow(["text", "company", "keywords", "relevant"])
        f.flush()

    batch_count = 0

    try:
        while rows_written["count"] < target_rows:
            try:
                row = await asyncio.wait_for(result_queue.get(), timeout=30)
            except asyncio.TimeoutError:
                continue

            writer.writerow([row.text, row.company, row.keywords, row.relevant])
            rows_written["count"] += 1
            batch_count += 1

            if batch_count % 100 == 0:
                f.flush()

            if batch_count % (CHECKPOINT_EVERY * COMPANIES_PER_TEXT) == 0:
                logging.info(f"Checkpoint: {rows_written['count']} rows written to disk")
    finally:
        f.flush()
        f.close()
        logging.info(f"Writer done. Total: {rows_written['count']}")


def print_stats(output_file: str):
    if not os.path.exists(output_file):
        return

    total = rel = irrel = 0
    by_company = {}
    unique_texts = set()

    with open(output_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            total += 1
            r = int(row["relevant"])
            rel += r
            irrel += (1 - r)
            unique_texts.add(row["text"][:100])

            comp = row["company"]
            if comp not in by_company:
                by_company[comp] = {"rel": 0, "irrel": 0}
            if r:
                by_company[comp]["rel"] += 1
            else:
                by_company[comp]["irrel"] += 1

    logging.info("=" * 70)
    logging.info("DATASET STATISTICS")
    logging.info(f"Total rows:    {total}")
    logging.info(f"Unique texts:  ~{len(unique_texts)}")
    logging.info(f"Relevant:      {rel} ({100*rel/max(1,total):.1f}%)")
    logging.info(f"Irrelevant:    {irrel} ({100*irrel/max(1,total):.1f}%)")
    logging.info("-" * 70)
    logging.info(f"{'Company':30s} {'Total':>6s} {'Rel':>6s} {'Irrel':>6s} {'Rel%':>6s}")
    logging.info("-" * 70)
    for comp in sorted(by_company):
        d = by_company[comp]
        t = d["rel"] + d["irrel"]
        pct = 100 * d["rel"] / max(1, t)
        logging.info(f"{comp:30s} {t:6d} {d['rel']:6d} {d['irrel']:6d} {pct:5.1f}%")
    logging.info("=" * 70)


async def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s [%(levelname)s] %(message)s",
        datefmt="%H:%M:%S",
    )

    logging.info(f"Relevance Dataset Generator v2")
    logging.info(f"Target: {TARGET_ROWS} rows | Workers: {MAX_WORKERS}")
    logging.info(f"Companies: {len(COMPANIES)} | Rows per text: {COMPANIES_PER_TEXT}")
    logging.info(f"Output: {OUTPUT_FILE}")

    existing = 0
    if os.path.exists(OUTPUT_FILE):
        with open(OUTPUT_FILE, "r", encoding="utf-8") as f:
            existing = max(0, sum(1 for _ in f) - 1)
        logging.info(f"Existing rows: {existing}")

    remaining = TARGET_ROWS - existing
    if remaining <= 0:
        logging.info("Target reached!")
        print_stats(OUTPUT_FILE)
        return

    logging.info(f"Generating {remaining} more rows (~{remaining // COMPANIES_PER_TEXT} texts)")

    queue = asyncio.Queue(maxsize=MAX_WORKERS * 20)
    counter = {"count": 0}
    written = {"count": existing}
    lock = asyncio.Lock()

    clients = []
    for _ in range(MAX_WORKERS):
        c = LLMClient(API_BASE_URL, API_KEY, MODEL_NAME)
        await c.init_session()
        clients.append(c)

    pbar = tqdm(total=remaining, desc="Generating rows", unit="row")

    tasks = [
        asyncio.create_task(worker(i, clients[i], queue, counter, remaining, lock, pbar))
        for i in range(MAX_WORKERS)
    ]

    writer_task = asyncio.create_task(
        csv_writer(queue, OUTPUT_FILE, remaining + existing, written)
    )

    t0 = time.time()
    await asyncio.gather(*tasks)
    
    pbar.close()

    while not queue.empty():
        await asyncio.sleep(0.5)
    await asyncio.sleep(2)

    writer_task.cancel()
    try:
        await writer_task
    except asyncio.CancelledError:
        pass

    for c in clients:
        await c.close()

    elapsed = time.time() - t0
    new = written["count"] - existing
    rate = new / elapsed if elapsed > 0 else 0

    logging.info(f"Done! New rows: {new} | Time: {elapsed:.0f}s | Rate: {rate:.1f} rows/sec")
    print_stats(OUTPUT_FILE)


if __name__ == "__main__":
    asyncio.run(main())
