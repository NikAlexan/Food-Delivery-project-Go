# 🍔 Food Delivery Service

> Микросервисный проект на Go — распределение задач по участникам команды.
> Деплой и конфиги - [ссылка](https://github.com/NikAlexan/sre-final)

---

## Архитектура

```
Client (Web / Mobile)
        │
        ▼
   API Gateway
        │
  ┌─────┼─────────┬──────────────┐
  ▼     ▼         ▼              ▼
User  Restaurant  Order       Delivery
Svc     Svc       Svc          Svc
  └─────┴─────────┴──────────────┘
                  │
              NATS JetStream
```

Все сервисы общаются между собой через **gRPC**. Асинхронные события — через **NATS JetStream**.

---

## Требования и баллы

| Требование | Баллы | Ответственный | Статус |
|---|:-:|---|:-:|
| Clean Architecture | 20% | Каждый студент | ✅ Выполнено |
| Минимум 12 gRPC Endpoints (3+ на сервис) | 20% | Каждый студент | ✅ Выполнено (34 endpoint'а) |
| NATS Message Queue | 20% | Nurassyl (Order) + интеграция | ✅ Выполнено |
| БД + миграции + транзакции | 20% | Каждый студент | ✅ Выполнено |
| Email через SMTP (Gmail / Microsoft) | 10% | Alikhan (Delivery) | ✅ Выполнено (MailHog) |
| Тесты (Unit + Integration) | 10% | Каждый студент | ✅ Выполнено |
| ⭐ Бонус: Frontend (JS / Native) | +10% | Nikita (по желанию) | ✅ Выполнено (Nuxt/Vue) |
| ⭐ Бонус: Grafana (трейсинг, метрики, логи) | +10% | Abzal (по желанию) | ✅ Выполнено |

---

## API Gateway — задача для всех

Каждый студент реализует Gateway для **своего** сервиса. Gateway принимает HTTP/REST и проксирует в gRPC.

| Маршрут | Сервис |
|---|---|
| `/api/users/*` | User Service |
| `/api/restaurants/*` | Restaurant Service |
| `/api/orders/*` | Order Service |
| `/api/delivery/*` | Delivery Service |

**Общие middleware:**

- JWT аутентификация
- Rate limiting
- Логирование запросов

**Стек:** `net/http` + `gorilla/mux` или `fiber` / `gin`

---

## 👤 Nikita Vassilenko — User Service

> Всё, что связано с пользователями: регистрация, вход, профиль, адреса.

### gRPC Endpoints

```
RegisterUser(RegisterRequest)         → UserResponse
LoginUser(LoginRequest)               → TokenResponse
GetProfile(UserIdRequest)             → UserProfile
UpdateProfile(UpdateProfileRequest)   → UserResponse
AddAddress(AddressRequest)            → AddressResponse
GetAddresses(UserIdRequest)           → AddressList
RefreshToken(RefreshRequest)          → TokenResponse
DeleteUser(UserIdRequest)             → Empty
```

### Структура (Clean Architecture)

```
user-service/
├── cmd/                  # точка входа gRPC сервера
├── internal/
│   ├── handler/          # gRPC хэндлеры (transport layer)
│   ├── usecase/          # бизнес-логика
│   ├── repository/       # работа с БД
│   └── model/            # доменные модели
├── proto/                # .proto файлы
└── migrations/           # SQL миграции (Goose / Migrate)
```

### База данных

- **PostgreSQL:** таблицы `users`, `addresses`, `refresh_tokens`
- **Redis:** кэш сессий и профилей
- **Миграции:** `create_users_table`, `add_addresses_table`
- **Транзакции:** создание `user` + `address` атомарно при регистрации

### NATS

| Действие | Топик |
|---|---|
| Публикует | `user.registered` |
| Подписывается | — |

### Тесты

- **Unit:** `usecase/register_test.go`, `usecase/login_test.go`
- **Integration:** тест `RegisterUser` через реальный PostgreSQL (testcontainers)

### ⭐ Бонус

- Frontend: страница регистрации и входа (React или Vue)

---

## 🍽️ Abzal Bakhtiyarov — Restaurant & Menu Service

> Рестораны, меню и каталог блюд. Опциональный бонус — Grafana.

### gRPC Endpoints

```
CreateRestaurant(CreateRestaurantRequest) → Restaurant
GetRestaurant(RestaurantIdRequest)        → Restaurant
ListRestaurants(ListRequest)              → RestaurantList
SearchRestaurants(SearchRequest)          → RestaurantList
CreateMenuItem(MenuItemRequest)           → MenuItem
UpdateMenuItem(UpdateMenuItemRequest)     → MenuItem
GetMenu(RestaurantIdRequest)              → MenuList
DeleteMenuItem(MenuItemIdRequest)         → Empty
```

### Структура (Clean Architecture)

```
restaurant-service/
├── cmd/
├── internal/
│   ├── handler/
│   ├── usecase/
│   ├── repository/
│   └── model/
├── proto/
└── migrations/
```

### База данных

- **PostgreSQL:** таблицы `restaurants`, `menu_items`, `categories`
- **Redis:** кэш меню (TTL 5 минут)
- **Миграции:** `create_restaurants_table`, `create_menu_items_table`
- **Транзакции:** batch update меню

### NATS

| Действие | Топик |
|---|---|
| Публикует | `restaurant.menu_updated` |
| Подписывается | — |

### Тесты

- **Unit:** `usecase/menu_test.go`, `usecase/search_test.go`
- **Integration:** тест `GetMenu` через реальный PostgreSQL

### ⭐ Бонус — Grafana Stack

- OpenTelemetry → **Tempo** (traces) + **Prometheus** (metrics) + **Loki** (logs)
- Подключить все 4 сервиса к общему Grafana стеку через `docker-compose`

---

## 🛒 Nurassyl Mukhtaruly — Order Service

> Ядро системы: создание заказов, обработка оплаты, публикация событий в NATS.

### gRPC Endpoints

```
CreateOrder(CreateOrderRequest)       → Order
GetOrder(OrderIdRequest)              → Order
ListUserOrders(UserIdRequest)         → OrderList
UpdateOrderStatus(UpdateStatusRequest)→ Order
CancelOrder(OrderIdRequest)           → Order
ProcessPayment(PaymentRequest)        → PaymentResponse
GetOrderHistory(UserIdRequest)        → OrderList
CalculateTotal(CartRequest)           → TotalResponse
```

### Структура (Clean Architecture)

```
order-service/
├── cmd/
├── internal/
│   ├── handler/
│   ├── usecase/
│   ├── repository/
│   ├── model/
│   └── nats/             # publisher и subscriber
├── proto/
└── migrations/
```

### База данных

- **PostgreSQL:** таблицы `orders`, `order_items`, `payments`
- **Redis:** кэш статусов заказов
- **Миграции:** `create_orders_table`, `create_payments_table`
- **Транзакции:** создание заказа + списание оплаты атомарно ⚠️

### NATS

| Действие | Топик |
|---|---|
| Публикует | `order.created`, `order.paid`, `order.cancelled` |
| Подписывается | — (инициатор событий) |

> Настроить **NATS JetStream** для надёжной доставки сообщений.

### Тесты

- **Unit:** `usecase/create_order_test.go`, `usecase/payment_test.go`
- **Integration:** тест полного цикла заказа с реальным PostgreSQL + NATS

---

## 🚴 Alikhan Maratbekov — Delivery Service + Email

> Трекинг доставки, назначение курьеров и отправка email-уведомлений через SMTP.

### gRPC Endpoints

```
AssignDriver(AssignRequest)             → Delivery
GetDelivery(DeliveryIdRequest)          → Delivery
UpdateDriverLocation(LocationRequest)   → Empty
TrackDelivery(DeliveryIdRequest)        → DeliveryStatus
CompleteDelivery(DeliveryIdRequest)     → Delivery
ListDriverDeliveries(DriverIdRequest)   → DeliveryList
GetDeliveryHistory(UserIdRequest)       → DeliveryList
CancelDelivery(DeliveryIdRequest)       → Delivery
```

### Структура (Clean Architecture)

```
delivery-service/
├── cmd/
├── internal/
│   ├── handler/
│   ├── usecase/
│   ├── repository/
│   ├── model/
│   ├── nats/             # подписчик на события заказов
│   └── email/            # SMTP клиент (Gmail OAuth / Microsoft)
├── proto/
└── migrations/
```

### База данных

- **PostgreSQL:** таблицы `deliveries`, `drivers`, `locations`
- **Redis:** кэш текущих локаций курьеров
- **Миграции:** `create_deliveries_table`, `create_drivers_table`
- **Транзакции:** проверка доступности курьера + назначение атомарно

### NATS

| Действие | Топик | Что делает |
|---|---|---|
| Подписывается | `order.created` | Назначить курьера |
| Подписывается | `order.paid` | Подтвердить доставку |
| Подписывается | `order.cancelled` | Освободить курьера |
| Публикует | `delivery.completed` | — |

### Email уведомления (SMTP)

| Триггер | Письмо |
|---|---|
| `order.created` | Подтверждение заказа |
| Delivery назначен | Заказ в пути |
| `delivery.completed` | Заказ доставлен |

**Реализация:** Gmail OAuth2 или Microsoft SMTP  
**Пакет:** `gopkg.in/gomail.v2` или `net/smtp`

### Тесты

- **Unit:** `usecase/assign_driver_test.go`, `email/smtp_test.go`
- **Integration:** тест назначения курьера + отправки email

---

## Технологический стек

| Слой | Технология |
|---|---|
| Язык | Go |
| Transport | gRPC (protobuf) |
| API Gateway | HTTP/REST → gRPC proxy |
| Message Queue | NATS JetStream |
| База данных | PostgreSQL |
| Кэш | Redis |
| Миграции | Goose / golang-migrate |
| Тесты | testcontainers-go |
| Observability (бонус) | OpenTelemetry + Grafana Stack |
| Frontend (бонус) | React / Vue / Native |
