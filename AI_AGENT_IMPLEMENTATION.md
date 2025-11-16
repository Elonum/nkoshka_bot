# Руководство по реализации AI агента

## Формат запросов от бота

Бот отправляет все запросы в едином формате:

```json
{
  "endpoint": "/endpoint_name",
  "data": {
    // Данные для обработки
  },
  "tg_id": 123456789,
  "timestamp": 1703520000
}
```

## Исправление кода AI агента

### Ваш текущий код (правильный для /api/auth/init)

Ваш код уже правильный для endpoint `/api/auth/init`:
```python
@app.post("/api/auth/init")
async def init_user(data: InitUserRequest):
    tg_id = data.tg_id
    username = data.username
    return {"status": "initialized", "tg_id": tg_id}
```

**Это правильно!** Бот отправляет данные напрямую в формате `InitUserRequest` для `/api/auth/init`.

### Для других endpoints нужна обёртка

Для всех остальных endpoints (генерация текста, изображений и т.д.) бот отправляет обёрнутый формат:

```python
from fastapi import FastAPI
from pydantic import BaseModel
from typing import Optional
import httpx

app = FastAPI()

# Модель для запроса от бота (для endpoints кроме /api/auth/init)
class BotRequest(BaseModel):
    endpoint: str
    data: dict
    tg_id: int
    timestamp: int

# Модель для запроса в бэкенд
class InitUserRequest(BaseModel):
    tg_id: int
    username: str

# URL бэкенда (из переменных окружения)
BACKEND_URL = "http://localhost:8080"  # или из env

# /api/auth/init - прямой формат (ваш код уже правильный)
@app.post("/api/auth/init")
async def init_user(data: InitUserRequest):
    """
    Обрабатывает запрос от бота для инициализации пользователя.
    Бот отправляет: {"tg_id": 123456789, "username": "test_user"}
    """
    tg_id = data.tg_id
    username = data.username
    
    # Если нужно вызвать бэкенд:
    # async with httpx.AsyncClient() as client:
    #     response = await client.post(
    #         f"{BACKEND_URL}/api/auth/init",
    #         json={"tg_id": tg_id, "username": username}
    #     )
    #     response.raise_for_status()
    
    return {"status": "initialized", "tg_id": tg_id}

@app.post("/local_bot/start")
async def bot_start(bot_request: BotRequest):
    """
    Дополнительный endpoint для старта бота (если нужен).
    """
    # Аналогично обрабатываем запрос от бота
    user_data = bot_request.data
    
    backend_request = InitUserRequest(
        tg_id=user_data.get("tg_id") or bot_request.tg_id,
        username=user_data.get("username", "")
    )
    
    # Вызываем бэкенд или выполняем нужную логику
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BACKEND_URL}/api/auth/init",
            json=backend_request.dict()
        )
        response.raise_for_status()
    
    return {"status": "started", "tg_id": backend_request.tg_id}
```

## Обработка других endpoints

Для всех остальных endpoints (генерация текста, изображений и т.д.) используйте аналогичный подход:

```python
@app.post("/generate_text")
async def generate_text(bot_request: BotRequest):
    """
    Обрабатывает запрос на генерацию текста.
    """
    # Извлекаем данные
    prompt = bot_request.data.get("prompt", "")
    nko_data = bot_request.data.get("nko", {})
    
    # Улучшаем промпт с учётом данных НКО (если нужно)
    enhanced_prompt = build_enhanced_prompt(prompt, nko_data)
    
    # Вызываем бэкенд
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BACKEND_URL}/api/tool/generate_text",
            json={"prompt": enhanced_prompt}
        )
        response.raise_for_status()
        result = response.json()
    
    # Возвращаем результат в формате PostJSON
    return {
        "post_id": result.get("post_id", ""),
        "post_author": bot_request.tg_id,
        "assigned_chat_id": [],
        "main_text": result.get("main_text", ""),
        "content": result.get("content", [])
    }

@app.post("/generate_image")
async def generate_image(bot_request: BotRequest):
    """
    Обрабатывает запрос на генерацию изображения.
    """
    # Извлекаем данные
    desc = bot_request.data.get("desc", "")
    nko_data = bot_request.data.get("nko", {})
    
    # Преобразуем desc в prompt
    prompt = build_image_prompt(desc, nko_data)
    
    # Вызываем бэкенд
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{BACKEND_URL}/api/tool/generate_image",
            json={"prompt": prompt, "aspect_ratio": "1x1"}
        )
        response.raise_for_status()
        result = response.json()
    
    # Возвращаем результат в формате PostJSON
    return {
        "post_id": result.get("post_id", ""),
        "post_author": bot_request.tg_id,
        "assigned_chat_id": [],
        "main_text": "",
        "content": result.get("content", [])
    }
```

## Вспомогательные функции

```python
def build_enhanced_prompt(prompt: str, nko_data: dict) -> str:
    """Улучшает промпт с учётом данных НКО."""
    enhanced = prompt
    
    if nko_data.get("name"):
        enhanced = f"НКО: {nko_data['name']}. {enhanced}"
    
    if nko_data.get("style"):
        style_instructions = {
            "разговорный": "Используй разговорный, живой стиль",
            "официальный": "Используй официальный, деловой стиль",
            # ... другие стили
        }
        if nko_data["style"] in style_instructions:
            enhanced = f"{style_instructions[nko_data['style']]}. {enhanced}"
    
    return enhanced

def build_image_prompt(desc: str, nko_data: dict) -> str:
    """Строит промпт для генерации изображения."""
    prompt = desc
    
    if nko_data.get("name"):
        prompt = f"Изображение для НКО '{nko_data['name']}': {prompt}"
    
    return prompt
```

## Важные замечания

1. **Всегда извлекайте данные из `bot_request.data`** - там находятся фактические данные запроса.

2. **Используйте `bot_request.tg_id`** как fallback, если `tg_id` не указан в `data`.

3. **Формат ответа** должен соответствовать `PostJSON` (см. `AI_AGENT_FORMAT.md`).

4. **Обработка ошибок:** Всегда обрабатывайте ошибки от бэкенда и возвращайте понятные сообщения боту.

5. **Логирование:** Логируйте все запросы и ответы для отладки.

## Пример полной структуры

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import httpx
import os
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI()

# Конфигурация
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")

# Модели
class BotRequest(BaseModel):
    endpoint: str
    data: dict
    tg_id: int
    timestamp: int

class InitUserRequest(BaseModel):
    tg_id: int
    username: str

# Endpoints
@app.post("/api/auth/init")
async def init_user(bot_request: BotRequest):
    try:
        user_data = bot_request.data
        backend_request = InitUserRequest(
            tg_id=user_data.get("tg_id") or bot_request.tg_id,
            username=user_data.get("username", "")
        )
        
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{BACKEND_URL}/api/auth/init",
                json=backend_request.dict(),
                timeout=10.0
            )
            response.raise_for_status()
        
        logger.info(f"User initialized: tg_id={backend_request.tg_id}")
        return {"status": "initialized", "tg_id": backend_request.tg_id}
    
    except httpx.HTTPError as e:
        logger.error(f"Backend error: {e}")
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# Добавьте остальные endpoints аналогично...
```

