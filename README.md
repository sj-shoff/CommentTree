# CommentTree - Система древовидных комментариев

## Описание

CommentTree - это сервис для работы с древовидными комментариями с неограниченной вложенностью. Система поддерживает создание, просмотр, удаление комментариев, полнотекстовый поиск, сортировку и пагинацию.

## Технологический стек

### Backend
- **Язык**: Go 1.24+
- **Фреймворк**: Chi Router
- **База данных**: PostgreSQL
- **Миграции**: Goose
- **Логирование**: Zerolog
- **Валидация**: go-playground/validator
- **Конфигурация**: cleanenv

### Frontend
- **HTML5**, **CSS3**, **JavaScript (ES6)**
- **Font Awesome** для иконок
- **Vanilla JavaScript** без фреймворков

### Инфраструктура
- **Docker** и **Docker Compose**
- **Makefile** для автоматизации

## Функции

### Основные функции
1. **Древовидные комментарии** - неограниченная вложенность комментариев
2. **CRUD операции** - создание, чтение, удаление комментариев
3. **Полнотекстовый поиск** - поиск по содержимому комментариев
4. **Сортировка** - по дате создания, обновления, ID
5. **Пагинация** - постраничный вывод комментариев
6. **Веб-интерфейс** - интуитивный UI для управления комментариями

### API Endpoints
- `POST /api/comments` - создание комментария
- `GET /api/comments` - получение комментариев с фильтрацией
- `DELETE /api/comments/{id}` - удаление комментария и всех дочерних

## Особенности

1. **Рекурсивное удаление** - при удалении комментария удаляются все дочерние
2. **Оптимизированные запросы** - использование рекурсивных CTE в PostgreSQL
3. **Построение дерева** - эффективное построение древовидной структуры в репозитории
4. **Полнотекстовый поиск** - поиск с использованием ILIKE
5. **Валидация данных** - проверка входных данных на стороне сервера
6. **Обработка ошибок** - структурированные ответы об ошибках
7. **SPA интерфейс** - одностраничное приложение без перезагрузок

## Быстрый старт

### Предварительные требования
- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)

### Запуск через Docker Compose

```bash
# 1. Клонируйте репозиторий
git clone <repository-url>
cd comments-system

# 2. Запустите приложение
make docker-up

# 3. Примените миграции
make migrate-up

# 4. Запустите приложение
make run

# 5. Проверьте работу
curl http://localhost:8080/api/health
```

## Примеры запросов

### Создание комментария

```bash
# Создать корневой комментарий
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Первый комментарий",
    "author": "Иван Иванов"
  }'

# Создать ответ на комментарий
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": 1,
    "content": "Ответ на первый комментарий",
    "author": "Петр Петров"
  }'
```

**Ответ:**
```json
{
  "id": 2,
  "parent_id": 1,
  "content": "Ответ на первый комментарий",
  "author": "Петр Петров",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Получение комментариев

```bash
# Получить все корневые комментарии с пагинацией
curl "http://localhost:8080/api/comments?page=1&page_size=10"

# Получить дерево комментариев для родителя
curl "http://localhost:8080/api/comments?parent=1"

# Поиск комментариев
curl "http://localhost:8080/api/comments?search=комментарий"

# Сортировка по дате создания
curl "http://localhost:8080/api/comments?sort_by=created_at&sort_order=desc"
```

**Ответ:**
```json
{
  "comments": [
    {
      "id": 1,
      "parent_id": null,
      "content": "Первый комментарий",
      "author": "Иван Иванов",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "children": [
        {
          "id": 2,
          "parent_id": 1,
          "content": "Ответ на комментарий",
          "author": "Петр Петров",
          "created_at": "2024-01-15T10:30:00Z",
          "updated_at": "2024-01-15T10:30:00Z",
          "children": []
        }
      ]
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10,
  "has_next": false,
  "has_prev": false
}
```

### Удаление комментария

```bash
# Удалить комментарий и все дочерние
curl -X DELETE http://localhost:8080/api/comments/1
```

**Ответ:** HTTP 204 No Content

## Структура проекта

```
comments-system/
├── cmd/comments-system/main.go     # Точка входа
├── internal/
│   ├── app/                        # Composition root
│   ├── config/                     # Конфигурация
│   ├── domain/                     # Доменные модели
│   ├── http-server/                # HTTP-сервер
│   │   ├── handler/                # Обработчики запросов
│   │   ├── middleware/             # Промежуточное ПО
│   │   └── router/                 # Маршрутизация
│   ├── repository/                 # Репозитории (PostgreSQL)
│   └── usecase/                    # Бизнес-логика
├── migrations/                     # Миграции базы данных
├── static/                         # Статические файлы (CSS, JS)
├── templates/                      # HTML шаблоны
├── docker-compose.yaml             # Docker Compose конфигурация
├── Makefile                        # Команды для разработки
├── go.mod                          # Зависимости Go
└── README.md                       # Документация
```

## Makefile команды

```bash
make docker-up      # Запуск приложения через Docker Compose
make docker-down    # Остановка контейнеров
make migrate-up     # Применение миграций базы данных
make migrate-down   # Откат миграций
make run           # Запуск приложения локально
make build         # Сборка приложения
```

## Конфигурация

Создайте файл `.env` на основе `.env.example`:

```env
# Server Configuration
SERVER_PORT=8004
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=60s
SERVER_SHUTDOWN_TIMEOUT=10s

# Database Configuration (PostgreSQL)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=comments
POSTGRES_MAX_OPEN_CONNS=10
POSTGRES_MAX_IDLE_CONNS=5
POSTGRES_CONN_MAX_LIFETIME=5m

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Retry Strategy
RETRIES_ATTEMPTS=3
RETRIES_DELAY_MS=2000
RETRIES_BACKOFF=2
```

## Миграции базы данных

База данных создается автоматически при первом запуске. Для ручного управления миграциями:

```bash
# Применить миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Показать статус миграций
goose -dir migrations postgres "postgres://comment_user:comment_password@localhost:5432/comment_system?sslmode=disable" status
```

## Веб-интерфейс

Веб-интерфейс доступен по адресу: http://localhost:8080

### Возможности интерфейса:
1. **Просмотр дерева** - визуальное отображение вложенности с отступами
2. **Создание комментариев** - форма с валидацией
3. **Ответы на комментарии** - кнопка "Ответить" для каждого комментария
4. **Удаление** - удаление с подтверждением
5. **Поиск** - мгновенный поиск по комментариям
6. **Пагинация** - навигация по страницам
7. **Сортировка** - переключение порядка сортировки

## Мониторинг и логи

- Логи приложения выводятся в формате JSON
- Доступны через `docker-compose logs -f app`
- Включено автоматическое восстановление после паники
- Логирование всех HTTP-запросов

## Производительность

- **Connection pooling** - пул соединений с базой данных
- **Рекурсивные запросы** - эффективное получение деревьев через CTE
- **Пагинация** - ограничение выборки для больших объемов данных
- **Индексы** - индексы на родительском ID и дате создания

## Ограничения

1. Максимальная длина имени автора: 50 символов
2. Максимальная длина комментария: 1000 символов
3. Максимальный размер страницы: 100 комментариев
4. Для поиска используется оператор ILIKE (регистронезависимый LIKE)
