#!/bin/bash

LOG_FILE="sentiment_train.log"

nohup uv run src/models/sentimental/train.py > "$LOG_FILE" 2>&1 < /dev/null &

echo $! > sentiment_train.pid

echo "Процесс запущен в фоне и будет продолжать работу после закрытия SSH соединения."
echo "Лог записывается в $LOG_FILE"
echo "Для остановки процесса выполните: kill $(cat sentiment_train.pid)"