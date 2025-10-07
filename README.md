Отлично! Создам comprehensive README.md для вашего RSS агрегатора. Вот полная документация:

# gator - RSS агрегатор новостных каналов

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![RSS](https://img.shields.io/badge/RSS-FFA500?style=for-the-badge&logo=rss&logoColor=white)

**gator** - это мощный CLI-инструмент для агрегации и управления RSS-фидами. Собирайте новости из любимых источников, организуйте подписки и читайте контент прямо в терминале.

## 🚀 Быстрый старт

### Предварительные требования

- **Go** 1.24 или выше [скачать](https://golang.org/dl/)
- **PostgreSQL** 12 или выше [скачать](https://www.postgresql.org/download/)

### Установка

```bash
# Установка из исходного кода
go install github.com/IdrisovMarat/agregator@latest

# Или клонирование и сборка
git clone https://github.com/IdrisovMarat/agregator
cd agregator
go build -o agregator .
```

### Настройка

1. **Создайте конфигурационный файл** `config.json`:

```json
{
  "db_url": "postgres://username:password@localhost:5432/agregator?sslmode=disable",
  "current_user_name": "your_username"
}
```

2. **Настройте базу данных**:

```bash
# Создайте базу данных
createdb agregator

# Или используйте существующую и обновите connection string в config.json
```

3. **Запустите миграции**:

```bash
agregator reset
```

## 📖 Использование

### Базовые команды

#### 🔐 Аутентификация
```bash
# Регистрация нового пользователя
agregator register <username>

# Вход пользователя
agregator login <username>

# Просмотр всех пользователей
agregator users
```

#### 📰 Управление фидами
```bash
# Добавление нового RSS фида
agregator addfeed "Название фида" "https://example.com/feed.xml"

# Просмотр всех фидов
agregator feeds

# Подписка на существующий фид
agregator follow "https://example.com/feed.xml"

# Просмотр ваших подписок
agregator following

# Отписка от фида
agregator unfollow <feed_id>
```

#### 👀 Просмотр контента
```bash
# Просмотр последних постов (по умолчанию 2)
agregator browse

# Просмотр N последних постов
agregator browse 10
```

#### 🔄 Автоматическая агрегация
```bash
# Запуск агрегатора с интервалом 1 минута
agregator agg 1m

# Запуск с другими интервалами
agregator agg 30s
agregator agg 5m
```

### Пример рабочего процесса

```bash
# 1. Регистрация и вход
agregator register marat
agregator login marat

# 2. Добавление фидов
agregator addfeed "Хабр" "https://habr.com/ru/rss/hub/go/all/"
agregator addfeed "TechCrunch" "https://techcrunch.com/feed/"

# 3. Запуск агрегатора в фоне (в одном терминале)
agregator agg 2m

# 4. Просмотр новостей (в другом терминале)
agregator browse 5
```

## 🗄️ Структура базы данных

gator использует следующую схему базы данных:

- **users** - пользователи системы
- **feeds** - RSS фиды с уникальными URL
- **feed_follows** - подписки пользователей на фиды
- **posts** - сохраненные посты из фидов

## ⚙️ Конфигурация

### Параметры config.json

| Параметр | Описание | Пример |
|----------|----------|---------|
| `db_url` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/agregator` |
| `current_user_name` | Текущий активный пользователь | `"marat"` |

### Переменные окружения

Вы также можете использовать переменные окружения:

```bash
export AGREGATOR_DB_URL="postgres://..."
export AGREGATOR_CURRENT_USER="username"
```

## 🛠️ Разработка

### Локальная разработка

```bash
# Клонирование репозитория
git clone https://github.com/IdrisovMarat/agregator
cd agregator

# Запуск в режиме разработки
go run .

# Запуск тестов
go test ./...
```

### Структура проекта

```
agregator/
├── internal/
│   ├── command/      # Обработчики команд CLI
│   ├── config/       # Конфигурация приложения
│   ├── database/     # Слой работы с базой данных
│   └── aggregator/   # Логика агрегации RSS
├── sql/
│   ├── queries/      # SQL запросы
│   └── migrations/   # Миграции базы данных
├── config.json       # Конфигурационный файл
└── README.md         # Документация
```

## 🔧 Миграции базы данных

Миграции выполняются автоматически при запуске команды `reset`:

```bash
agregator reset
```

Или вручную с помощью Goose:

```bash
goose -dir sql/migrations postgres "your-db-url" up
```

## 📊 Поддерживаемые RSS форматы

Agregator поддерживает:
- RSS 2.0
- Атом (Atom)
- Различные форматы дат публикации
- HTML entities в заголовках и описаниях

## 🚨 Решение проблем

### Частые проблемы

**Ошибка подключения к базе данных**
```bash
# Проверьте connection string в config.json
# Убедитесь, что PostgreSQL запущен
sudo service postgresql start
```

**Фиды не обновляются**
```bash
# Убедитесь, что агрегатор запущен
agregator agg 1m

# Проверьте URL фида
agregator feeds
```

**Дубликаты постов**
- Это нормальное поведение - программа автоматически игнорирует дубликаты

## 🤝 Вклад в развитие

Мы приветствуем вклады в развитие проекта!

1. Форкните репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Запушьте branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. Подробнее см. в файле `LICENSE`.

## 👨‍💻 Автор

**Marat Idrisov**
- GitHub: [@IdrisovMarat](https://github.com/IdrisovMarat)

## 🌟 Особенности

- ✅ Автоматическая агрегация RSS фидов
- ✅ Управление подписками
- ✅ Чтение в терминале
- ✅ Поддержка множества пользователей
- ✅ Защита от DOS-атак
- ✅ Обработка ошибок и дубликатов
- ✅ Красивый форматированный вывод

---

**Наслаждайтесь чтением новостей без отвлечений!** 📰✨