# REST API Organization Structure

Сервис на Go для управления древовидной структурой департаментов и сотрудниками.

## Стек технологий
* **Язык:** Go 1.24+
* **База данных:** PostgreSQL
* **Миграции:** goose
* **ORM:** GORM
* **Конфигурация:** godotenv (локально) / Docker Environment
* **Роутинг:** Стандартный net/http (без сторонних фреймворков)
* **Контейнеризация:** Docker (Multi-stage build), Docker Compose

## Ключевая логика
* **Детекция циклов:** BFS-обход графа блокирует попытки сделать департамент дочерним для самого себя или своих потомков.
* **Инверсия зависимостей:** Сервисы зависят от интерфейсов репозиториев — бизнес-логика изолирована и покрыта Unit-тестами (Table-driven паттерн).
* **Логирование:** Middleware логирует HTTP-запросы (метод, URL, время обработки).

---

## Запуск

```bash
docker-compose up --build
```

API доступен на `http://localhost:8080`. Миграции применяются автоматически.

### Taskfile
* `task up` — поднять контейнеры
* `task down` — остановить и очистить
* `task test` — запустить Unit-тесты

---

## API

### Департаменты

#### POST /api/departments — создать департамент
```json
{ "name": "Разработка", "parent_id": null }
```
Ответ `201`:
```json
{ "id": 1, "name": "Разработка", "parent_id": null }
```

#### GET /api/departments — дерево всех департаментов
Query: `?with_employees=true`

#### GET /api/departments/{id} — получить департамент
Query:
* `depth` — глубина вложенных департаментов (1–5, по умолчанию 1)
* `include_employees` — включить сотрудников (по умолчанию true)

#### PATCH /api/departments/{id} — обновить департамент
```json
{ "name": "Новое название", "parent_id": 3 }
```
Ответ при цикле `400`:
```json
{ "error": "cyclical dependency detected: cannot set sub-department as parent" }
```

#### DELETE /api/departments/{id} — удалить департамент
Query:
* `mode` — `cascade` (по умолчанию) или `reassign`
* `reassign_to_department_id` — обязателен при `mode=reassign`

Ответ `204 No Content`

---

### Сотрудники

#### POST /api/departments/{id}/employees — создать сотрудника в департаменте
```json
{ "full_name": "Иван Иванов", "position": "Go Developer", "hired_at": "2024-01-15" }
```
Ответ `201`:
```json
{ "id": 1, "department_id": 2, "full_name": "Иван Иванов", "position": "Go Developer", "hired_at": "2024-01-15" }
```