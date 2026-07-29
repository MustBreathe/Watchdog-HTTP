# 📊 Watchdog-HTTP

Aplikacja do monitorowania dostępności stron i API w Go.

## O Projekcie

**Watchdog HTTP** — monitor dostępności stron i API. Zbuduj w Go aplikację, która:
- 🔄 Cyklicznie sprawdza dostępność wskazanych adresów HTTP/HTTPS
- 💾 Zapisuje wyniki w bazie danych
- 📡 Wystawia REST API
- 🎛️ Pozwala zarządzać monitorami przez CLI
- 🔔 Wysyła powiadomienia webhookiem, gdy usługa przestaje działać

---

## 🚀 Quick Start

### Uruchomienie serwera
```bash
go run ./cmd/watchdog/main.go server
```

### Instalacja narzędzia do migracji
```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Po instalacji `migrate.exe` będzie dostępny w: `C:\Users\LukasiM\go\bin`

---
---

## 📋 Spis Treści

1. [Cel projektu](#cel-projektu)
2. [Zakres techniczny](#zakres-techniczny)
3. [Proponowany stack](#proponowany-stack)
4. [Wymagania funkcjonalne](#wymagania-funkcjonalne)
   - [Zarządzanie monitorami](#zarządzanie-monitorami)
   - [Wykonywanie sprawdzeń HTTP](#wykonywanie-sprawdzeń-http)
   - [Scheduler](#scheduler)
   - [Worker pool](#worker-pool)
5. [REST API](#rest-api)
6. [CLI](#cli)
7. [Baza danych](#baza-danych)
8. [Powiadomienia webhook](#powiadomienia-webhook)
9. [Konfiguracja aplikacji](#konfiguracja-aplikacji)
10. [Logowanie](#logowanie)
11. [Struktura katalogów](#struktura-katalogów)
12. [Testy](#testy)
13. [Etapy realizacji](#etapy-realizacji)

---

## 🎯 Cel projektu
Aplikacja ma pozwalać użytkownikowi zdefiniować listę monitorów, np.:

| Pole | Wartość |
|------|---------|
| Nazwa | API produkcyjne |
| URL | https://example.com/health |
| Metoda | GET |
| Interwał | 60 sekund |
| Timeout | 5 sekund |
| Oczekiwany status | 200 |
| Oczekiwany tekst | "ok" |

**System ma automatycznie:**
- Sprawdzać te adresy
- Zapisywać historię wyników
- Umożliwiać sprawdzenie, które usługi działają, które nie działają
- Wyświetlać dostępność w czasie

---

## ⚙️ Zakres techniczny

Aplikacja powinna składać się z czterech głównych części:

1. **REST API** — do zarządzania monitorami i odczytu wyników
2. **Scheduler + worker pool** — do wykonywania cyklicznych sprawdzeń HTTP
3. **Baza danych** — do przechowywania monitorów, wyników i zdarzeń
4. **CLI** — do zarządzania aplikacją z terminala

Opcjonalnie: prosty panel HTML (nie obowiązkowy)

---

## 📦 Proponowany stack

### Minimalnie wymagane

- Go 1.22+
- `net/http`
- `database/sql`
- SQLite albo PostgreSQL
- `encoding/json`
- `context`
- `testing`

### Opcjonalne biblioteki

```
github.com/mattn/go-sqlite3 albo modernc.org/sqlite  — SQLite driver
github.com/jackc/pgx/v5                              — PostgreSQL driver
github.com/spf13/cobra                               — CLI framework
github.com/joho/godotenv                             — zmienne środowiskowe
github.com/google/uuid                               — UUID
```

⚠️ **Wskazówka:** Nie musisz używać frameworka webowego. Dla treningu Go lepiej zacząć od `net/http`.

---

## 📝 Wymagania funkcjonalne

### Zarządzanie monitorami

System musi pozwalać na:
- ➕ Tworzenie monitorów
- ✏️ Edycję monitorów
- 🗑️ Usuwanie monitorów
- ✅ Włączanie i wyłączanie
- 📋 Listowanie monitorów

#### Pola monitora

Każdy monitor powinien mieć następujące pola:

```
- id                         — UUID
- name                       — nazwa monitora
- url                        — adres URL
- method                     — metoda HTTP (GET, POST, itd.)
- interval_seconds           — interwał sprawdzania
- timeout_seconds            — timeout żądania
- expected_status            — spodziewany kod HTTP
- expected_body_substring    — fragment odpowiedzi (opcjonalnie)
- headers                    — dodatkowe nagłówki
- enabled                    — czy monitor jest aktywny
- created_at, updated_at     — znaczniki czasowe
```

#### Walidacja

```json
{
  "name": "Example API",
  "url": "https://example.com/health",
  "method": "GET",
  "interval_seconds": 60,
  "timeout_seconds": 5,
  "expected_status": 200,
  "expected_body_substring": "ok",
  "headers": {
    "Authorization": "Bearer test-token"
  },
  "enabled": true
}
```

**Reguły walidacji:**

| Pole | Reguła |
|------|--------|
| `name` | Wymagane, 3-100 znaków |
| `url` | Wymagane, musi zaczynać się od `http://` lub `https://` |
| `method` | Tylko: GET, POST, PUT, PATCH, DELETE, HEAD |
| `interval_seconds` | 10–86400 |
| `timeout_seconds` | 1–60 |
| `expected_status` | 100–599 |
| `expected_body_substring` | Opcjonalne, max 500 znaków |
| `headers` | Opcjonalne, max 20 nagłówków |

**Błąd walidacji:**

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "interval_seconds must be between 10 and 86400"
  }
}
```

### Wykonywanie sprawdzeń HTTP

Aplikacja ma cyklicznie wykonywać sprawdzenia dla wszystkich aktywnych monitorów.

#### Dane wyniku sprawdzenia

```
- id                       — UUID
- monitor_id               — ID monitora
- started_at, finished_at  — znaczniki czasu
- duration_ms              — czas odpowiedzi w ms
- status_code              — otrzymany kod HTTP
- success                  — czy sprawdzenie się powiodło
- error_message            — komunikat błędu (jeśli dostępny)
- response_body_sample     — fragment odpowiedzi
- created_at               — czas zapisania wyniku
```

#### Zasada sukcesu

Sprawdzenie jest uznane za udane, gdy:
- ✅ Żądanie HTTP zakończyło się bez błędu
- ✅ Kod statusu jest równy `expected_status`
- ✅ Jeśli ustawiony `expected_body_substring` — odpowiedź go zawiera
- ✅ Czas odpowiedzi nie przekroczył `timeout_seconds`

#### Przykład wyniku udanego sprawdzenia

```json
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "status_code": 200,
  "success": true,
  "duration_ms": 123,
  "error_message": null
}
```

#### Przykład wyniku nieudanego sprawdzenia

```json
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "status_code": 500,
  "success": false,
  "duration_ms": 88,
  "error_message": "expected status 200, got 500"
}
```

### Scheduler

Scheduler ma uruchamiać sprawdzenia zgodnie z interwałem każdego monitora.

**Wymagania:**

| Wymóg | Opis |
|-------|------|
| Regularność | Monitor z `interval_seconds = 60` ma być sprawdzany co 60 sekund |
| Ignorowanie wyłączonych | Monitor wyłączony nie może być sprawdzany |
| Dynamiczne zmiany | Zmiana interwału monitora bez restartu aplikacji |
| Usunięte monitory | Usunięty monitor nie może być dalej sprawdzany |
| Brak równoległych | Jeśli poprzednie sprawdzenie trwa, nie uruchamiaj kolejnego |
| Graceful shutdown | Reaguj na `context cancellation` przy zamykaniu |

> ⚠️ **Ważne:** To jedna z kluczowych części projektu. Wymusi użycie goroutines, contextów, synchronizacji i kanałów.

### Worker pool

Nie uruchamiaj nieograniczonej liczby goroutines. Zaimplementuj worker pool.

**Konfiguracja:**

```bash
WORKER_COUNT=5
CHECK_QUEUE_SIZE=100
```

**Wymagania:**

- 📬 Scheduler wrzuca zadania sprawdzeń do kolejki
- 👷 Pracownicy pobierają zadania z kanału
- 🔍 Każdy worker wykonuje sprawdzenie HTTP
- 💾 Wynik jest zapisywany do bazy
- ⚠️ Jeśli kolejka jest pełna: zaloguj ostrzeżenie, pomiń zadanie lub zwróć błąd 500

**Model zadania:**

```go
type CheckJob struct {
    MonitorID string  // ID monitora
    Trigger   string  // "scheduled" albo "manual"
}
```

---

## 🔌 REST API

API powinno działać domyślnie na porcie `8080`.

### Format błędów

Wszystkie błędy powinny mieć jeden format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "monitor not found"
  }
}
```

**Kody błędów:**

- `VALIDATION_ERROR` — błąd walidacji
- `NOT_FOUND` — zasób nie znaleziony
- `UNAUTHORIZED` — brak autoryzacji
- `INTERNAL_ERROR` — błąd serwera
- `CONFLICT` — konflikt (np. pełna kolejka)
### Autoryzacja

Dodaj prostą autoryzację przez API key.

Każde żądanie do API (poza `/healthz`) musi zawierać nagłówek:

```
X-API-Key: secret-dev-key
```

Klucz ma być pobierany z konfiguracji:

```bash
API_KEY=secret-dev-key
```

**Jeśli klucz jest błędny:**

```http
401 Unauthorized
```

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid api key"
  }
}
```

### Endpointy techniczne

#### `GET /healthz`

Zwraca status aplikacji. **Nie wymaga API key.**

**Odpowiedź:**

```json
{
  "status": "ok"
}
```

### Endpointy monitorów

#### `POST /api/v1/monitors` — Utwórz monitor

Tworzy nowy monitor.

**Request:**

```json
{
  "name": "Example API",
  "url": "https://example.com/health",
  "method": "GET",
  "interval_seconds": 60,
  "timeout_seconds": 5,
  "expected_status": 200,
  "expected_body_substring": "ok",
  "headers": {
    "Authorization": "Bearer token"
  },
  "enabled": true
}
```

**Response** `201 Created`:

```json
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "name": "Example API",
  "url": "https://example.com/health",
  "method": "GET",
  "interval_seconds": 60,
  "timeout_seconds": 5,
  "expected_status": 200,
  "expected_body_substring": "ok",
  "headers": {
    "Authorization": "Bearer token"
  },
  "enabled": true,
  "created_at": "2026-06-16T12:00:00Z",
  "updated_at": "2026-06-16T12:00:00Z"
}
```

#### `GET /api/v1/monitors` — Lista monitorów

Zwraca listę monitorów.

**Parametry query:**

```
enabled=true|false   — opcjonalne
limit=50             — domyślnie 50, maksymalnie 200
offset=0             — domyślnie 0
```

**Response:**

```json
{
  "items": [
    {
      "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
      "name": "Example API",
      "url": "https://example.com/health",
      "method": "GET",
      "interval_seconds": 60,
      "timeout_seconds": 5,
      "expected_status": 200,
      "enabled": true,
      "created_at": "2026-06-16T12:00:00Z",
      "updated_at": "2026-06-16T12:00:00Z"
    }
  ],
  "limit": 50,
  "offset": 0,
  "total": 1
}
```

#### `GET /api/v1/monitors/{id}` — Szczegóły monitora

Zwraca szczegóły monitora.

**Response** `200 OK`:

```json
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "name": "Example API",
  "url": "https://example.com/health",
  "method": "GET",
  "interval_seconds": 60,
  "timeout_seconds": 5,
  "expected_status": 200,
  "expected_body_substring": "ok",
  "headers": {
    "Authorization": "Bearer token"
  },
  "enabled": true,
  "created_at": "2026-06-16T12:00:00Z",
  "updated_at": "2026-06-16T12:00:00Z"
}
```

**Jeśli monitor nie istnieje:**

```http
404 Not Found
```

#### `PUT /api/v1/monitors/{id}` — Aktualizuj monitor

Aktualizuje cały monitor.

- **Request:** taki sam jak przy tworzeniu
- **Response** `200 OK`: zaktualizowany monitor

#### `PATCH /api/v1/monitors/{id}/enable` — Włącz monitor
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "enabled": true
**Response:**

```json
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "enabled": true
}
```

#### `PATCH /api/v1/monitors/{id}/disable` — Wyłącz monitor

**Response:**

```json
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "enabled": false
}
```

#### `DELETE /api/v1/monitors/{id}` — Usuń monitor

**Response** `204 No Content`

> 💡 Usunięcie monitora powinno również usunąć jego wyniki. Prostsze podejście: fizyczne usuwanie.

### Endpointy sprawdzeń

#### `POST /api/v1/monitors/{id}/checks` — Ręczne sprawdzenie

Uruchamia ręczne sprawdzenie monitora.

**Response** `202 Accepted`:

```json
{
  "message": "check queued",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31"
}
```

**Jeśli kolejka jest pełna:**

```http
409 Conflict
```

```json
{
  "error": {
    "code": "CONFLICT",
    "message": "check queue is full"
  }
}
```

#### `GET /api/v1/monitors/{id}/checks` — Historia sprawdzeń

Zwraca historię sprawdzeń danego monitora.

**Parametry:**

```
limit=50
offset=0
success=true|false   (opcjonalnie)
```

**Response:**

```json
{
  "items": [
    {
      "id": "b8b8d71d-7a88-4639-bce0-d3e9bb731a12",
      "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
      "started_at": "2026-06-16T12:00:00Z",
      "finished_at": "2026-06-16T12:00:00Z",
      "duration_ms": 123,
      "status_code": 200,
      "success": true,
      "error_message": null,
      "response_body_sample": "ok"
    }
  ],
  "limit": 50,
  "offset": 0,
  "total": 1
}
```

#### `GET /api/v1/checks/{id}` — Szczegóły sprawdzenia

Zwraca pojedynczy wynik sprawdzenia.

**Response:**

```json
{
  "id": "b8b8d71d-7a88-4639-bce0-d3e9bb731a12",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "started_at": "2026-06-16T12:00:00Z",
  "finished_at": "2026-06-16T12:00:00Z",
  "duration_ms": 123,
  "status_code": 200,
  "success": true,
  "error_message": null,
  "response_body_sample": "ok"
}
```

### Endpoint statusu monitora

#### `GET /api/v1/monitors/{id}/status` — Status monitora

Zwraca aktualny status monitora obliczony na podstawie ostatniego sprawdzenia.

**Response:**

```json
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "name": "Example API",
  "current_status": "up",
  "last_checked_at": "2026-06-16T12:00:00Z",
  "last_duration_ms": 123,
  "last_status_code": 200,
  "last_error_message": null
}
```

**Możliwe wartości** `current_status`:

- `up` — ostatnie sprawdzenie było udane
- `down` — ostatnie sprawdzenie było nieudane
- `unknown` — monitor nie ma jeszcze żadnych sprawdzeń

### Endpoint statystyk

#### `GET /api/v1/monitors/{id}/stats` — Statystyki monitora

Zwraca statystyki dla monitora.

**Parametry:**

```
from=2026-06-01T00:00:00Z   (opcjonalnie)
to=2026-06-16T23:59:59Z     (opcjonalnie)
```

**Response:**

```json
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "total_checks": 100,
  "successful_checks": 95,
  "failed_checks": 5,
  "uptime_percentage": 95.0,
  "average_duration_ms": 142,
  "min_duration_ms": 80,
  "max_duration_ms": 410
}
```

**Wyjaśnienie pól:**

| Pole | Opis |
|------|------|
| `total_checks` | Liczba wszystkich sprawdzeń w zakresie |
| `successful_checks` | Liczba udanych sprawdzeń |
| `failed_checks` | Liczba nieudanych sprawdzeń |
| `uptime_percentage` | Procent czasu działania (successful/total * 100) |
| `average_duration_ms` | Średni czas odpowiedzi |
| `min_duration_ms` | Najmniejszy czas odpowiedzi |
| `max_duration_ms` | Największy czas odpowiedzi |

---

## 💻 CLI

Zbuduj prostą aplikację CLI, która komunikuje się z REST API.

**Struktura:**

```bash
watchdog server                # Uruchomienie serwera
watchdog monitor list          # Lista monitorów
watchdog monitor add           # Dodanie monitora
watchdog monitor get <ID>      # Szczegóły monitora
watchdog monitor delete <ID>   # Usunięcie monitora
watchdog monitor check <ID>    # Ręczne sprawdzenie
watchdog monitor status <ID>   # Status monitora
watchdog monitor checks <ID>   # Historia sprawdzeń
```

### Konfiguracja CLI

CLI powinno czytać zmienne środowiskowe:
--api-key
Flagi powinny mieć pierwszeństwo nad zmiennymi środowiskowymi.
6.2. Wymagane komendy CLI
Uruchomienie serwera
watchdog server
Uruchamia REST API, scheduler i worker pool.
Dodanie monitora
watchdog monitor add \
  --name "Example API" \
  --url "https://example.com/health" \
  --method GET \
  --interval 60 \
  --timeout 5 \
  --expected-status 200 \
  --expected-body "ok"
Po sukcesie:
Monitor created:
ID: 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
Name: Example API
URL: https://example.com/health
Lista monitorów
```bash
WATCHDOG_API_URL=http://localhost:8080
WATCHDOG_API_KEY=secret-dev-key
```

**Flagi opcjonalne:**

```bash
--api-url <URL>      # Nadpisz URL API
--api-key <KEY>      # Nadpisz klucz API
```

Flagi mają pierwszeństwo nad zmiennymi środowiskowymi.

### Wymagane komendy CLI

#### `watchdog server` — Uruchomienie serwera

Uruchamia REST API, scheduler i worker pool.

```bash
watchdog server
```

#### `watchdog monitor add` — Dodanie monitora

```bash
watchdog monitor add \
  --name "Example API" \
  --url "https://example.com/health" \
  --method GET \
  --interval 60 \
  --timeout 5 \
  --expected-status 200 \
  --expected-body "ok"
```

**Po sukcesie:**

```
Monitor created:
ID: 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
Name: Example API
URL: https://example.com/health
```

#### `watchdog monitor list` — Lista monitorów

```bash
watchdog monitor list
```

**Output:**

```
ID                                    NAME          URL                         STATUS   INTERVAL
8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31  Example API   https://example.com/health  up       60s
```

#### `watchdog monitor get <ID>` — Szczegóły monitora

```bash
watchdog monitor get 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
```

#### `watchdog monitor check <ID>` — Ręczne sprawdzenie

```bash
watchdog monitor check 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
```

**Po sukcesie:**

```
Check queued.
```

#### `watchdog monitor status <ID>` — Status monitora

```bash
watchdog monitor status 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
```

**Output:**

```
Status: up
Last checked: 2026-06-16 12:00:00
Duration: 123 ms
HTTP status: 200
```

#### `watchdog monitor checks <ID>` — Historia sprawdzeń

```bash
watchdog monitor checks 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31 --limit 10
```

**Output:**

```
TIME                  SUCCESS   STATUS   DURATION   ERROR
2026-06-16 12:00:00   true      200      123ms      -
2026-06-16 11:59:00   false     500      88ms       expected status 200, got 500
```

#### `watchdog monitor delete <ID>` — Usunięcie monitora

```bash
watchdog monitor delete 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
```

CLI powinno poprosić o potwierdzenie:

```
Are you sure? Type monitor name to confirm:
```

Alternatywnie dodaj flagę `--yes` dla pominięcia potwierdzenia.

---

## 🗄️ Baza danych

Możesz użyć SQLite, żeby projekt był łatwy do uruchomienia lokalnie. Schemat powinien być zarządzany przez migracje SQL.

### Tabela `monitors`

```sql
CREATE TABLE monitors (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL,
    timeout_seconds INTEGER NOT NULL,
    expected_status INTEGER NOT NULL,
    expected_body_substring TEXT,
    headers_json TEXT,
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

### Tabela `checks`

```sql
CREATE TABLE checks (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL,
    duration_ms INTEGER NOT NULL,
    status_code INTEGER,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    response_body_sample TEXT,
    trigger TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);
```

**Indeksy:**

```sql
CREATE INDEX idx_checks_monitor_id_created_at
ON checks (monitor_id, created_at DESC);

CREATE INDEX idx_checks_success
ON checks (success);
```

### Tabela `incidents`

Dodaj ją w drugiej części projektu, po zrobieniu podstawowego monitoringu.

```sql
CREATE TABLE incidents (
    id TEXT PRIMARY KEY,
    monitor_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    resolved_at TEXT,
    status TEXT NOT NULL,
    failure_count INTEGER NOT NULL DEFAULT 1,
    last_error_message TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);
```

**Możliwe wartości** `status`: `open`, `resolved`

**Zasada:**

- 🔴 Jeśli monitor przejdzie ze `up` lub `unknown` → `down`: utwórz incident
- ➕ Jeśli monitor jest dalej `down`: zwiększ `failure_count`
- 🟢 Jeśli monitor przejdzie z `down` → `up`: zamknij incident (`resolved_at`, `status=resolved`)

---

## 🔔 Powiadomienia webhook

Po wykryciu awarii aplikacja powinna wysłać webhook na skonfigurowany adres.

**Konfiguracja:**

```bash
WEBHOOK_URL=https://example.com/webhook
WEBHOOK_ENABLED=true
```

**Webhook wysyłany przy zmianie stanu:**

```
unknown → down
up → down
down → up
```

> ⚠️ Nie wysyłaj webhooka przy każdym nieudanym sprawdzeniu, jeśli monitor już jest w stanie `down`.

### Payload webhooka

**Dla awarii** (`monitor_down`):

```json
{
  "event": "monitor_down",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "monitor_name": "Example API",
  "url": "https://example.com/health",
  "checked_at": "2026-06-16T12:00:00Z",
  "error_message": "expected status 200, got 500"
}
```

**Dla powrotu do działania** (`monitor_up`):

```json
{
  "event": "monitor_up",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "monitor_name": "Example API",
  "url": "https://example.com/health",
  "checked_at": "2026-06-16T12:05:00Z"
}
```

### Obsługa błędów webhooka

| Wymóg | Opis |
|-------|------|
| Timeout | Webhook ma timeout **5 sekund** |
| Kody błędu | Jeśli status poza 200-299: zaloguj błąd |
| Nie przerywaj | Webhook fail nie wstrzymuje monitorowania |
| Context timeout | Użyj timeout dla worker pool |

---

## ⚙️ Konfiguracja aplikacji

Aplikacja powinna wspierać konfigurację przez zmienne środowiskowe.

**Wymagane zmienne:**

```bash
APP_ENV=dev
HTTP_ADDR=:8080
DATABASE_DSN=./watchdog.db
API_KEY=secret-dev-key
WORKER_COUNT=5
CHECK_QUEUE_SIZE=100
SCHEDULER_TICK_SECONDS=5
WEBHOOK_ENABLED=false
WEBHOOK_URL=
LOG_LEVEL=info
```

**Reguły walidacji:**

- Zmienne z sensowną domyślną wartością mogą być opcjonalne
- `API_KEY` w produkcji nie może być puste
- `WORKER_COUNT` > 0
- `CHECK_QUEUE_SIZE` > 0
- `HTTP_ADDR` musi być poprawnym adresem

---

## 📝 Logowanie

Dodaj uporządkowane logi. Możesz użyć standardowego `log/slog`.

**Logi powinny zawierać:**

```
✅ Start aplikacji
✅ Załadowana konfiguracja (bez sekretów!)
✅ Uruchomienie serwera HTTP
✅ Start i stop schedulera
✅ Start i stop workerów
✅ Każde wykonane sprawdzenie
✅ Błędy HTTP clienta
✅ Błędy zapisu do bazy
✅ Wysłane webhooki
✅ Błędne requesty API
```

**Przykładowy log:**

```
level=INFO msg="check completed" monitor_id=8f4b5d2e status_code=200 success=true duration_ms=123
```

⚠️ **Nigdy nie loguj:**
- Wartości `API_KEY`
- Nagłówków `Authorization`
- Sensytywnych danych

---

## 📁 Struktura katalogów

Proponowana struktura (nie musisz być sztywny):

```
watchdog/
  ├── cmd/
  │   └── watchdog/
  │       ├── main.go
  │       ├── cli.go
  │       └── runner.go
  ├── internal/
  │   ├── api/
  │   │   ├── handlers.go
  │   │   ├── middleware.go
  │   │   ├── routes.go
  │   │   └── errors.go
  │   ├── app/
  │   │   └── app.go
  │   ├── checker/
  │   │   ├── checker.go
  │   │   └── result.go
  │   ├── config/
  │   │   └── config.go
  │   ├── monitor/
  │   │   ├── model.go
  │   │   ├── repository.go
  │   │   ├── service.go
  │   │   └── validation.go
  │   ├── scheduler/
  │   │   ├── scheduler.go
  │   │   └── worker_pool.go
  │   ├── incident/
  │   │   ├── model.go
  │   │   ├── repository.go
  │   │   └── service.go
  │   ├── notify/
  │   │   └── webhook.go
  │   └── storage/
  │       ├── db.go
  │       └── migrations.go
  ├── migrations/
  │   ├── 001_create_monitors.sql
  │   ├── 002_create_checks.sql
  │   └── 003_create_incidents.sql
  ├── tests/
  │   └── integration/
  ├── go.mod
  ├── README.md
  ├── Dockerfile
  ├── docker-compose.yml
  └── Makefile
```

> 💡 Warto rozdzielić logikę domenową od handlerów HTTP.

---

## 🏗️ Podział na warstwy

### Handler HTTP

Odpowiada **tylko** za:
- Odczyt requestu
- Walidacja podstawowego JSON-a
- Wywołanie serwisu
- Zwrócenie response

### Service

Odpowiada za logikę biznesową:
- Tworzenie monitora
- Walidacja reguł
- Włączanie/wyłączanie
- Obliczanie statusu
- Uruchamianie ręcznego sprawdzenia
- Obsługę incidentów

### Repository

Odpowiada za bazę danych:
- `InsertMonitor`, `GetMonitor`, `ListMonitors`, `UpdateMonitor`, `DeleteMonitor`
- `InsertCheck`, `ListChecks`, `GetStats`

### Checker

Odpowiada za wykonanie pojedynczego sprawdzenia HTTP:

```go
type Checker interface {
    Check(ctx context.Context, monitor Monitor) CheckResult
}
```

### Scheduler

Odpowiada za planowanie zadań:

```go
type Scheduler interface {
    Start(ctx context.Context) error
}
```

---

## 🧪 Testy
- zwrócenie response.
Handler nie powinien sam wykonywać SQL-a ani logiki schedulera.
Service
Odpowiada za logikę biznesową:
- tworzenie monitora,
- walidację reguł,
- włączanie/wyłączanie,
- obliczanie statusu,
- uruchamianie ręcznego sprawdzenia,
- obsługę incidentów.
Repository
Odpowiada za bazę danych:
- InsertMonitor
- GetMonitor
- ListMonitors
- UpdateMonitor
- DeleteMonitor
- InsertCheck
- ListChecks
- GetStats
---

## 🧪 Testy

To powinien być projekt z realnymi testami, nie tylko ręczne klikanie.

### Testy jednostkowe

**Wymagane testy:**

- ✅ Walidacja monitora
- ✅ Budowanie requestu HTTP
- ✅ Interpretacja odpowiedzi HTTP
- ✅ Timeout requestu
- ✅ `expected_status` — czy działa poprawnie
- ✅ `expected_body_substring` — czy działa poprawnie
- ✅ Obliczanie statystyk uptime
- ✅ Formatowanie błędów API
- ✅ Autoryzacja API key
- ✅ Logika incidentów: down → incident, kolejne down → licznik, up → zamknięcie

**Do testów HTTP użyj** `httptest.Server`.

**Przykładowe scenariusze:**

```
1. Serwer testowy zwraca 200 i body "ok"          → check success ✅
2. Serwer testowy zwraca 500                       → check failed ❌
3. Serwer testowy zwraca 200, ale bez tekstu       → check failed ❌
4. Serwer testowy śpi dłużej niż timeout           → check failed ❌
```

### Testy integracyjne

**Minimum:**

- ✅ Utworzenie monitora przez API
- ✅ Pobranie monitora przez API
- ✅ Ręczne uruchomienie sprawdzenia
- ✅ Zapis wyniku w bazie
- ✅ Pobranie historii sprawdzeń
- ✅ Usunięcie monitora

> 💡 Test integracyjny może uruchamiać aplikację z tymczasową bazą SQLite.

### Pokrycie testami

**Cel:** minimum **60% coverage** dla `internal/`

**Komenda:**

```bash
go test ./... -cover
```

---

## 🛑 Obsługa zamykania aplikacji

Aplikacja musi poprawnie reagować na `SIGINT` i `SIGTERM`.

**Po naciśnięciu Ctrl+C:**

- 🛑 Serwer HTTP przestaje przyjmować nowe requesty
- 🛑 Scheduler przestaje dodawać nowe zadania
- ⏱️ Workerzy kończą aktualnie wykonywane zadania (lub przerywają przez context timeout)
- 🗄️ Połączenie z bazą danych zostaje zamknięte
- ✅ Aplikacja kończy się bez panic

> 💡 To ćwiczenie nauczy Cię poprawnego użycia `context.Context`.

---

## 🐳 Docker i uruchamianie

### Dockerfile

**Wymagania:**

- ✅ Build aplikacji w osobnym etapie (multi-stage)
- ✅ Finalny obraz zawiera tylko binarkę i migracje
- ✅ Aplikacja uruchamia się: `watchdog server`

### docker-compose.yml

Przykład z SQLite:

```yaml
services:
  watchdog:
    build: .
    ports:
      - "8080:8080"
    environment:
      HTTP_ADDR: ":8080"
      DATABASE_DSN: "/data/watchdog.db"
      API_KEY: "secret-dev-key"
    volumes:
      - watchdog_data:/data

volumes:
  watchdog_data:
```

---

## 🛠️ Makefile

Dodaj podstawowe komendy:

```makefile
.PHONY: run test cover lint build docker-build

run:
	go run ./cmd/watchdog server

test:
	go test ./...

cover:
	go test ./... -cover

lint:
	go vet ./...

build:
	go build -o bin/watchdog ./cmd/watchdog

docker-build:
	docker build -t watchdog .
```

---

## 📖 README

README powinno zawierać:

- ✅ Opis projektu
- ✅ Wymagania systemowe
- ✅ Instrukcję uruchomienia lokalnego
- ✅ Instrukcję uruchomienia przez Docker
- ✅ Przykłady użycia API (curl)
- ✅ Przykłady użycia CLI
- ✅ Opis zmiennych środowiskowych
- ✅ Opis architektury
- ✅ Znane ograniczenia

**Przykład curl do utworzenia monitora:**

```bash
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-dev-key" \
  -d '{
    "name": "Example API",
    "url": "https://example.com/health",
    "method": "GET",
    "interval_seconds": 60,
    "timeout_seconds": 5,
    "expected_status": 200,
    "expected_body_substring": "ok",
    "enabled": true
  }'
```

---

## 📊 Etapy realizacji
- przykłady użycia CLI,
- opis zmiennych środowiskowych,
- opis architektury,
- opis decyzji projektowych,
- opis znanych ograniczeń.
Przykład curl do utworzenia monitora:
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-dev-key" \
  -d '{
    "name": "Example API",
    "url": "https://example.com/health",
    "method": "GET",
    "interval_seconds": 60,
    "timeout_seconds": 5,
    "expected_status": 200,
    "expected_body_substring": "ok",
    "enabled": true
  }'
```

---

## 📊 Etapy realizacji

### Etap 1 — Szkielet projektu

**Zakres:**
- ✅ Moduł Go
- ✅ Struktura katalogów
- ✅ Konfiguracja
- ✅ Logger
- ✅ Endpoint `GET /healthz`
- ✅ Makefile

**Kryterium ukończenia:**

```bash
go run ./cmd/watchdog server
curl http://localhost:8080/healthz
# {"status":"ok"}
```

### Etap 2 — Baza danych i monitory

**Zakres:**
- ✅ Migracje SQL
- ✅ Tabela `monitors`
- ✅ Repository dla monitorów
- ✅ Endpointy CRUD
- ✅ Walidacja danych
- ✅ Testy walidacji

**Kryterium:** Można CRUD-ować monitory przez REST API.

### Etap 3 — Checker HTTP

**Zakres:**
- ✅ Wykonywanie sprawdzenia
- ✅ Timeout (context)
- ✅ `expected_status`
- ✅ `expected_body_substring`
- ✅ Zapis wyniku
- ✅ Testy z httptest

**Kryterium:** `POST /api/v1/monitors/{id}/checks` działa.

### Etap 4 — Worker pool

**Zakres:**
- ✅ Kolejka zadań
- ✅ Workerzy
- ✅ Obsługa pełnej kolejki
- ✅ Graceful shutdown

**Kryterium:** Sprawdzenia działają przez worker pool.

### Etap 5 — Scheduler

**Zakres:**
- ✅ Cykliczne sprawdzenia
- ✅ Respektowanie interwałów
- ✅ Brak duplikatów
- ✅ Dynamiczne zmiany

**Kryterium:** Monitor sprawdza się automatycznie.

### Etap 6 — Status i statystyki

**Zakres:**
- ✅ Endpoint statusu
- ✅ Endpoint statystyk
- ✅ Obliczanie uptime
- ✅ Historia

**Kryterium:** Można sprawdzić status i uptime.

### Etap 7 — Incidenty i webhooki

**Zakres:**
- ✅ Tabela incidents
- ✅ Przejścia up ↔ down
- ✅ Wysyłanie webhooka
- ✅ Testy

**Kryterium:** Webhook wysyłany tylko przy zmianie stanu.

### Etap 8 — CLI

**Zakres:**
- ✅ Wszystkie komendy
- ✅ Konfiguracja
- ✅ Komunikaty błędów

**Kryterium:** CLI w pełni funkcjonalny.

### Etap 9 — Docker & dokumentacja

**Zakres:**
- ✅ Dockerfile
- ✅ docker-compose.yml
- ✅ README
- ✅ Przykłady
- ✅ Testy
- ✅ Coverage

**Kryterium:** Nowa osoba może uruchomić projekt z README.

---

## ✅ Definicja ukończenia

Projekt gotów, gdy spełnia wszystkie warunki:

1. ✅ Uruchamia się lokalnie jedną komendą
2. ✅ Można dodać monitor
3. ✅ Monitor sprawdza się automatycznie
4. ✅ Wyniki w bazie
5. ✅ Historia dostępna przez API
6. ✅ Status dostępny
7. ✅ Statystyki uptime dostępne
8. ✅ Ręczne sprawdzenia przez worker pool
9. ✅ Graceful shutdown
10. ✅ API key wymagany
11. ✅ CLI pełny
12. ✅ `go test ./...` przechodzi
13. ✅ README kompletny
14. ✅ Docker działający

---

## 🚀 Dodatkowe wyzwania

Po ukończeniu głównej wersji:

- 🎨 Panel HTML (html/template)
- 📊 Export CSV
- 📈 Prometheus metrics
- 🏥 Endpoint /readyz
- 🔄 Retry policy
- 🌐 TCP monitoring
- ⚡ Rate limiting
- 👥 Role użytkowników
- 📢 Slack, Discord, e-mail
- 📚 OpenAPI/Swagger
- ⏱️ Benchmarki
- 🧪 Property-based tests
- 💾 Import/export

---

## 🎓 Czego się nauczysz

**Go Essentials:**
- 🔌 REST API (net/http)
- 📝 JSON (encoding/json)
- 🗄️ Baza danych (database/sql)
- 🎯 Context, goroutines, channels
- 👷 Worker pool
- 🛑 Graceful shutdown

**Quality:**
- ✅ Testy jednostkowe i integracyjne
- 🧪 httptest
- 📊 Coverage
- 🔍 Logging (slog)

**DevOps:**
- 🐳 Docker & docker-compose
- 📋 Makefile
- ⚙️ Zmienne środowiskowe

---

## 💡 Rekomendowana kolejność pracy

```
REST API + Baza → Checker → Worker Pool → Scheduler
```

