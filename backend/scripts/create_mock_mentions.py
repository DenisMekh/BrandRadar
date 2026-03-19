#!/usr/bin/env python3
"""
Программа для массового создания тестовых упоминаний через /api/v1/mock/mention endpoint.
Создаёт 200 упоминаний с разницей в 30 секунд по PUBLISHED_AT:
  - Первые 100: очень позитивные отзывы о Т-Банке (~50 слов)
  - Следующие 50: негативные отзывы о Т-Банке (~50 слов)
  - Последние 50: снова позитивные отзывы (~50 слов)

Для использования измени КОНСТАНТЫ в начале файла и запусти:
    python scripts/create_mock_mentions.py
"""

import requests
import time
import uuid
from datetime import datetime, timedelta

# =============================================================================
# НАСТРОЙКИ — измени значения под свои нужды
# =============================================================================

BRAND_ID = "9f40d139-8af0-4b12-925a-7a62d7d4997f"  # UUID бренда
SOURCE_ID = "858ef5f9-1797-414a-b03f-40133cab09ca"  # UUID источника
COUNT = 200  # Общее количество упоминаний (100 + 50 + 50)
DELAY = 1  # Задержка между запросами API в секундах
TIME_DELTA = DELAY  # Разница в секундах между PUBLISHED_AT соседних упоминаний
BASE_URL = "http://localhost:8080"  # Базовый URL API
AUTHOR = "@testuser"  # Автор упоминания
START_TIME = None  # Начальное время (None = сейчас), формат: datetime

# =============================================================================
# ТЕКСТЫ ОТЗЫВОВ
# =============================================================================

POSITIVE_TEXT = """
Я очень доволен обслуживанием в Т-Банке! Это лучший банк, с которым мне приходилось иметь дело. 
Отличное мобильное приложение, быстрые переводы, выгодные условия по кредитам и депозитам. 
Кэшбэк возвращают регулярно, бонусы начисляют честно. Сотрудники всегда вежливые и готовые помочь. 
Поддержка работает круглосуточно, решают любые вопросы за считанные минуты. 
Рекомендую всем своим друзьям и знакомым! Т-Банк — это надёжность, качество и отличный сервис. 
Спасибо за прекрасное обслуживание и индивидуальный подход к каждому клиенту!
"""

NEGATIVE_TEXT = """
Ужасное обслуживание в Т-Банке! Постоянные проблемы с приложением, переводы не доходят. 
Кэшбэк не возвращают уже третий месяц, на звонки не отвечают. Сотрудники грубят и хамят. 
Кредит навязали с огромными процентами, условия изменили в одностороннем порядке. 
Поддержка игнорирует жалобы, проблемы не решают неделями. Комиссии скрытые, условия запутанные. 
Деньги заблокировали без предупреждения, объяснений никаких. Никому не рекомендую этот банк! 
Очень жалею, что открыл здесь счёт. Ищите нормальный банк с человеческим отношением!
"""

# =============================================================================


def get_text_for_index(index: int) -> str:
    """Возвращает текст в зависимости от индекса."""
    if index < 100:
        # Первые 100 — позитивные
        return POSITIVE_TEXT.strip()
    elif index < 150:
        # Следующие 50 — негативные
        return NEGATIVE_TEXT.strip()
    else:
        # Последние 50 — снова позитивные
        return POSITIVE_TEXT.strip()


def create_mention(
    base_url: str,
    brand_id: str,
    source_id: str,
    text: str,
    published_at: str,
    title: str = None,
    author: str = None,
) -> dict:
    """Создаёт одно упоминание через mock endpoint."""
    url = f"{base_url}/api/v1/mock/mention"
    
    payload = {
        "brand_id": brand_id,
        "source_id": source_id,
        "external_id": f"mock-{uuid.uuid4()}",
        "title": title or f"Test mention #{uuid.uuid4().hex[:8]}",
        "text": text,
        "url": f"https://t.me/test/{uuid.uuid4()}",
        "author": author or "@testuser",
        "published_at": published_at,
    }
    
    response = requests.post(url, json=payload, timeout=10)
    return {
        "status_code": response.status_code,
        "response": response.json() if response.status_code == 201 else response.text,
    }


def main():
    # Определяем начальное время
    base_time = START_TIME if START_TIME else datetime.utcnow()
    
    print(f"🚀 Запуск создания {COUNT} упоминаний...")
    print(f"   Brand ID: {BRAND_ID}")
    print(f"   Source ID: {SOURCE_ID}")
    print(f"   Time delta between mentions: {TIME_DELTA}s")
    print(f"   Base URL: {BASE_URL}")
    print(f"   Start time: {base_time.isoformat()}Z")
    print()
    print("📋 План:")
    print(f"   1-100:  Позитивные отзывы о Т-Банке")
    print(f"   101-150: Негативные отзывы о Т-Банке")
    print(f"   151-200: Позитивные отзывы о Т-Банке")
    print()

    success_count = 0
    error_count = 0

    for i in range(COUNT):
        # Вычисляем время публикации для этого упоминания
        mention_time = base_time + timedelta(seconds=i * TIME_DELTA)
        published_at = mention_time.strftime("%Y-%m-%dT%H:%M:%SZ")
        
        # Получаем текст в зависимости от индекса
        text = get_text_for_index(i)
        
        # Определяем тип отзыва для лога
        if i < 100:
            sentiment_type = "😊 POSITIVE"
        elif i < 150:
            sentiment_type = "😠 NEGATIVE"
        else:
            sentiment_type = "😊 POSITIVE"

        try:
            result = create_mention(
                base_url=BASE_URL,
                brand_id=BRAND_ID,
                source_id=SOURCE_ID,
                text=text,
                published_at=published_at,
                author=AUTHOR,
            )

            if result["status_code"] == 201:
                success_count += 1
                sentiment = result["response"].get("data", {}).get("sentiment_ml", {})
                ml_label = sentiment.get('label', 'N/A')
                ml_score = sentiment.get('score', 'N/A')
                print(f"✅ [{i+1:3d}/{COUNT}] {sentiment_type} | Published: {published_at} | ML: {ml_label} ({ml_score})")
            else:
                error_count += 1
                print(f"❌ [{i+1:3d}/{COUNT}] {sentiment_type} | Error {result['status_code']}: {result['response']}")

        except requests.exceptions.RequestException as e:
            error_count += 1
            print(f"❌ [{i+1:3d}/{COUNT}] {sentiment_type} | Request failed: {e}")

        if i < COUNT - 1:
            time.sleep(DELAY)

    print()
    print("=" * 70)
    print(f"📊 Результаты:")
    print(f"   Успешно: {success_count}")
    print(f"   Ошибки: {error_count}")
    print(f"   Всего: {COUNT}")
    print(f"   Время начала: {base_time.strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"   Время конца: {(base_time + timedelta(seconds=(COUNT-1)*TIME_DELTA)).strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 70)


if __name__ == "__main__":
    main()
