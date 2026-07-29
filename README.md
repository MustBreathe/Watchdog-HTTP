# Watchdog-HTTP

 Run Server
 ```
 go run .\cmd\watchdog\main.go server
 ```
 

## Migration
```
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
 Po instalacji migrate.exe będzie w:
C:\Users\LukasiM\go\bin

Projekt: Watchdog HTTP — monitor dostępności stron i API
Zbuduj w Go aplikację, która cyklicznie sprawdza dostępność wskazanych adresów HTTP/HTTPS, zapisuje wyniki w bazie danych, wystawia REST API, pozwala zarządzać monitorami przez CLI i wysyła powiadomienia webhookiem, gdy usługa przestaje działać.
To jest projekt na około 40–70 godzin, zależnie od poziomu dopracowania. Bardzo dobrze pasuje do Go, bo użyjesz praktycznie: net/http, context, goroutines, channels, worker pool, database/sql, testów, middleware, konfiguracji, logowania, obsługi błędów i prostego deployu.
1. Cel projektu
Aplikacja ma pozwalać użytkownikowi zdefiniować listę monitorów, np.:
Nazwa: API produkcyjne
URL: https://example.com/health
Metoda: GET
Interwał: 60 sekund
Timeout: 5 sekund
Oczekiwany status: 200
Oczekiwany tekst w odpowiedzi: "ok"
System ma automatycznie sprawdzać te adresy, zapisywać historię wyników i umożliwiać sprawdzenie, które usługi działają, które nie działają i jaka była ich dostępność w czasie.
2. Zakres techniczny
Aplikacja powinna składać się z czterech głównych części:
REST API do zarządzania monitorami i odczytu wyników.
Scheduler + worker pool do wykonywania cyklicznych sprawdzeń HTTP.
Baza danych do przechowywania monitorów, wyników i zdarzeń.
CLI do zarządzania aplikacją z terminala.
Opcjonalnie możesz dodać prosty panel HTML, ale nie jest konieczny.
3. Proponowany stack
Minimalnie:
Go 1.22+
net/http
database/sql
SQLite albo PostgreSQL
encoding/json
context
testing
Dodatkowo możesz użyć:
github.com/mattn/go-sqlite3 albo modernc.org/sqlite
github.com/jackc/pgx/v5 — jeżeli wybierzesz PostgreSQL
github.com/spf13/cobra — CLI
github.com/joho/godotenv — konfiguracja lokalna
github.com/google/uuid — UUID
Nie musisz używać frameworka webowego. Dla treningu Go lepiej zacząć od net/http.
4. Główne wymagania funkcjonalne
4.1. Zarządzanie monitorami
System musi pozwalać na tworzenie, edycję, usuwanie, włączanie, wyłączanie i listowanie monitorów HTTP.
Każdy monitor powinien mieć pola:
id
name
url
method
interval_seconds
timeout_seconds
expected_status
expected_body_substring
headers
enabled
created_at
updated_at
Wymagania szczegółowe
Użytkownik może utworzyć monitor z następującymi danymi:
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
Walidacja:
name: wymagane, od 3 do 100 znaków
url: wymagane, musi zaczynać się od http:// albo https://
method: tylko GET, POST, PUT, PATCH, DELETE, HEAD
interval_seconds: od 10 do 86400
timeout_seconds: od 1 do 60
expected_status: od 100 do 599
expected_body_substring: opcjonalne, maksymalnie 500 znaków
headers: opcjonalne, maksymalnie 20 nagłówków
enabled: boolean
Jeżeli dane są niepoprawne, API musi zwrócić błąd w ustalonym formacie:
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "interval_seconds must be between 10 and 86400"
  }
}
4.2. Wykonywanie sprawdzeń HTTP
Aplikacja ma cyklicznie wykonywać sprawdzenia dla wszystkich aktywnych monitorów.
Dla każdego sprawdzenia system powinien zapisać:
id
monitor_id
started_at
finished_at
duration_ms
status_code
success
error_message
response_body_sample
created_at
Zasada sukcesu
Sprawdzenie jest uznane za udane, jeżeli:
- żądanie HTTP zakończyło się bez błędu,
- kod statusu jest równy expected_status,
- jeżeli expected_body_substring jest ustawione, odpowiedź zawiera ten tekst,
- czas odpowiedzi nie przekroczył timeout_seconds.
Przykład wyniku udanego sprawdzenia:
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "status_code": 200,
  "success": true,
  "duration_ms": 123,
  "error_message": null
}
Przykład wyniku nieudanego sprawdzenia:
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "status_code": 500,
  "success": false,
  "duration_ms": 88,
  "error_message": "expected status 200, got 500"
}
4.3. Scheduler
Scheduler ma uruchamiać sprawdzenia zgodnie z interwałem każdego monitora.
Wymagania:
- monitor z interval_seconds = 60 ma być sprawdzany mniej więcej co 60 sekund,
- monitor wyłączony nie może być sprawdzany,
- zmiana interwału monitora powinna zostać uwzględniona bez restartu aplikacji,
- usunięty monitor nie może być dalej sprawdzany,
- jeżeli poprzednie sprawdzenie monitora jeszcze trwa, scheduler nie powinien uruchamiać kolejnego dla tego samego monitora,
- scheduler musi reagować na context cancellation przy zamykaniu aplikacji.
To jest jedna z najważniejszych części projektu, bo wymusi użycie goroutines, contextów, synchronizacji i kanałów.
4.4. Worker pool
Nie uruchamiaj nieograniczonej liczby goroutines. Zaimplementuj worker pool.
Konfiguracja:
WORKER_COUNT=5
CHECK_QUEUE_SIZE=100
Wymagania:
- scheduler wrzuca zadania sprawdzeń do kolejki,
- pracownicy pobierają zadania z kanału,
- każdy worker wykonuje sprawdzenie HTTP,
- wynik jest zapisywany do bazy,
- jeżeli kolejka jest pełna, system powinien zalogować ostrzeżenie i pominąć zadanie albo zwrócić błąd wewnętrzny — wybierz jedną strategię i opisz ją w README.
Przykładowy model zadania:
type CheckJob struct {
    MonitorID string
    Trigger   string // "scheduled" albo "manual"
}
5. REST API
API powinno działać domyślnie na porcie 8080.
5.1. Format błędów
Wszystkie błędy powinny mieć jeden format:
{
  "error": {
    "code": "NOT_FOUND",
    "message": "monitor not found"
  }
}
Przykładowe kody błędów:
VALIDATION_ERROR
NOT_FOUND
UNAUTHORIZED
INTERNAL_ERROR
CONFLICT
5.2. Autoryzacja
Dodaj prostą autoryzację przez API key.
Każde żądanie do API, poza /healthz, musi zawierać nagłówek:
X-API-Key: secret-dev-key
Klucz ma być pobierany z konfiguracji:
API_KEY=secret-dev-key
Jeżeli klucz jest błędny albo pusty, API zwraca:
401 Unauthorized
oraz:
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid api key"
  }
}
5.3. Endpointy techniczne
GET /healthz
Zwraca status aplikacji.
Odpowiedź:
{
  "status": "ok"
}
Ten endpoint nie wymaga API key.
5.4. Endpointy monitorów
POST /api/v1/monitors
Tworzy nowy monitor.
Request:
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
Response 201 Created:
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
GET /api/v1/monitors
Zwraca listę monitorów.
Parametry query:
enabled=true|false — opcjonalne
limit=50 — domyślnie 50, maksymalnie 200
offset=0 — domyślnie 0
Response:
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
GET /api/v1/monitors/{id}
Zwraca szczegóły monitora.
Response 200 OK:
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
Jeżeli monitor nie istnieje:
404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "monitor not found"
  }
}
PUT /api/v1/monitors/{id}
Aktualizuje cały monitor.
Request taki sam jak przy tworzeniu.
Response 200 OK: zaktualizowany monitor.
PATCH /api/v1/monitors/{id}/enable
Włącza monitor.
Response:
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "enabled": true
}
PATCH /api/v1/monitors/{id}/disable
Wyłącza monitor.
Response:
{
  "id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "enabled": false
}
DELETE /api/v1/monitors/{id}
Usuwa monitor.
Response:
204 No Content
Wymaganie: usunięcie monitora powinno usunąć również jego wyniki albo oznaczyć monitor jako usunięty. Wybierz jedną strategię. Na start prostsze jest fizyczne usuwanie monitora i wyników.
5.5. Endpointy sprawdzeń
POST /api/v1/monitors/{id}/checks
Uruchamia ręczne sprawdzenie monitora.
Response 202 Accepted:
{
  "message": "check queued",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31"
}
Jeżeli kolejka jest pełna:
409 Conflict
{
  "error": {
    "code": "CONFLICT",
    "message": "check queue is full"
  }
}
GET /api/v1/monitors/{id}/checks
Zwraca historię sprawdzeń danego monitora.
Parametry:
limit=50
offset=0
success=true|false — opcjonalne
Response:
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
GET /api/v1/checks/{id}
Zwraca pojedynczy wynik sprawdzenia.
Response:
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
5.6. Endpoint statusu monitora
GET /api/v1/monitors/{id}/status
Zwraca aktualny status monitora obliczony na podstawie ostatniego sprawdzenia.
Response:
{
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "name": "Example API",
  "current_status": "up",
  "last_checked_at": "2026-06-16T12:00:00Z",
  "last_duration_ms": 123,
  "last_status_code": 200,
  "last_error_message": null
}
Możliwe wartości current_status:
up
down
unknown
Zasady:
up: ostatnie sprawdzenie było udane
down: ostatnie sprawdzenie było nieudane
unknown: monitor nie ma jeszcze żadnych sprawdzeń
5.7. Endpoint statystyk
GET /api/v1/monitors/{id}/stats
Zwraca statystyki dla monitora.
Parametry:
from=2026-06-01T00:00:00Z — opcjonalne
to=2026-06-16T23:59:59Z — opcjonalne
Response:
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
Wymagania:
- total_checks = liczba wszystkich sprawdzeń w zakresie,
- successful_checks = liczba udanych sprawdzeń,
- failed_checks = liczba nieudanych sprawdzeń,
- uptime_percentage = successful_checks / total_checks * 100,
- average_duration_ms = średni czas odpowiedzi,
- jeżeli brak danych, uptime_percentage powinno wynosić null albo 0 — wybierz jedną strategię i opisz ją w README.
6. CLI
Zbuduj prostą aplikację CLI, która komunikuje się z REST API.
Możesz zrobić osobny binary:
watchdogctl
Albo jeden program z subkomendami:
watchdog server
watchdog monitor list
watchdog monitor add
6.1. Konfiguracja CLI
CLI powinno czytać:
WATCHDOG_API_URL=http://localhost:8080
WATCHDOG_API_KEY=secret-dev-key
Można też dodać flagi:
--api-url
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
watchdog monitor list
Przykładowy output:
ID                                    NAME          URL                         STATUS   INTERVAL
8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31  Example API   https://example.com/health  up       60s
Szczegóły monitora
watchdog monitor get 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
Ręczne sprawdzenie
watchdog monitor check 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
Po sukcesie:
Check queued.
Status monitora
watchdog monitor status 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
Output:
Status: up
Last checked: 2026-06-16 12:00:00
Duration: 123 ms
HTTP status: 200
Historia sprawdzeń
watchdog monitor checks 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31 --limit 10
Output:
TIME                  SUCCESS   STATUS   DURATION   ERROR
2026-06-16 12:00:00   true      200      123ms      -
2026-06-16 11:59:00   false     500      88ms       expected status 200, got 500
Usunięcie monitora
watchdog monitor delete 8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31
CLI powinno poprosić o potwierdzenie:
Are you sure? Type monitor name to confirm:
Dla prostszej wersji możesz dodać flagę:
--yes
7. Baza danych
Możesz użyć SQLite, żeby projekt był łatwy do uruchomienia lokalnie. Schemat powinien być zarządzany przez migracje SQL.
7.1. Tabela monitors
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
7.2. Tabela checks
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
Indeksy:
CREATE INDEX idx_checks_monitor_id_created_at
ON checks (monitor_id, created_at DESC);

CREATE INDEX idx_checks_success
ON checks (success);
7.3. Tabela incidents
Dodaj ją w drugiej części projektu, po zrobieniu podstawowego monitoringu.
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
Możliwe wartości status:
open
resolved
Zasada:
- jeżeli monitor przejdzie ze stanu up/unknown do down, utwórz incident,
- jeżeli monitor jest dalej down, zwiększ failure_count w otwartym incydencie,
- jeżeli monitor przejdzie z down do up, zamknij otwarty incident przez ustawienie resolved_at i status = resolved.
8. Powiadomienia webhook
Po wykryciu awarii aplikacja powinna wysłać webhook na skonfigurowany adres.
Konfiguracja:
WEBHOOK_URL=https://example.com/webhook
WEBHOOK_ENABLED=true
Webhook powinien być wysyłany tylko przy zmianie stanu:
unknown -> down
up -> down
down -> up
Nie wysyłaj webhooka przy każdym nieudanym sprawdzeniu, jeżeli monitor już jest w stanie down.
8.1. Payload webhooka
Dla awarii:
{
  "event": "monitor_down",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "monitor_name": "Example API",
  "url": "https://example.com/health",
  "checked_at": "2026-06-16T12:00:00Z",
  "error_message": "expected status 200, got 500"
}
Dla powrotu do działania:
{
  "event": "monitor_up",
  "monitor_id": "8f4b5d2e-8122-4c0d-b1ab-55c98f8a0c31",
  "monitor_name": "Example API",
  "url": "https://example.com/health",
  "checked_at": "2026-06-16T12:05:00Z"
}
8.2. Obsługa błędów webhooka
Wymagania:
- webhook ma timeout 5 sekund,
- jeżeli webhook zwróci status spoza zakresu 200–299, zaloguj błąd,
- nie przerywaj głównego sprawdzania monitorów, jeżeli webhook się nie uda,
- nie blokuj workerów zbyt długo — użyj context timeout.
9. Konfiguracja aplikacji
Aplikacja powinna wspierać konfigurację przez zmienne środowiskowe.
Wymagane zmienne:
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
Wymagania:
- jeżeli zmienna ma sensowną wartość domyślną, aplikacja może jej użyć,
- API_KEY w trybie produkcyjnym nie może być puste,
- WORKER_COUNT musi być większe od 0,
- CHECK_QUEUE_SIZE musi być większe od 0,
- HTTP_ADDR musi być poprawnym adresem dla serwera HTTP.
10. Logowanie
Dodaj uporządkowane logi. Możesz użyć standardowego log/slog.
Logi powinny zawierać:
- start aplikacji,
- załadowaną konfigurację bez sekretów,
- uruchomienie serwera HTTP,
- start i stop schedulera,
- start i stop workerów,
- każde wykonane sprawdzenie,
- błędy HTTP clienta,
- błędy zapisu do bazy,
- wysłane webhooki,
- błędne requesty API.
Przykładowy log:
level=INFO msg="check completed" monitor_id=8f4b5d2e status_code=200 success=true duration_ms=123
Nie loguj wartości API_KEY ani nagłówków typu Authorization.
11. Struktura katalogów
Proponowana struktura:
watchdog/
  cmd/
    watchdog/
      main.go
  internal/
    api/
      handlers.go
      middleware.go
      routes.go
      errors.go
    app/
      app.go
    checker/
      checker.go
      result.go
    config/
      config.go
    monitor/
      model.go
      repository.go
      service.go
      validation.go
    scheduler/
      scheduler.go
      worker_pool.go
    incident/
      model.go
      repository.go
      service.go
    notify/
      webhook.go
    storage/
      db.go
      migrations.go
  migrations/
    001_create_monitors.sql
    002_create_checks.sql
    003_create_incidents.sql
  tests/
    integration/
  go.mod
  README.md
  Dockerfile
  docker-compose.yml
  Makefile
Nie musisz trzymać się tego idealnie, ale warto rozdzielić logikę domenową od handlerów HTTP.
12. Minimalny podział na warstwy
Handler HTTP
Odpowiada tylko za:
- odczyt requestu,
- walidację podstawowego JSON-a,
- wywołanie serwisu,
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
Checker
Odpowiada za wykonanie jednego sprawdzenia HTTP:
type Checker interface {
    Check(ctx context.Context, monitor Monitor) CheckResult
}
Scheduler
Odpowiada za planowanie zadań:
type Scheduler interface {
    Start(ctx context.Context) error
}
13. Testy
To powinien być projekt z realnymi testami, nie tylko ręczne klikanie.
13.1. Testy jednostkowe
Wymagane testy:
- walidacja monitora,
- budowanie requestu HTTP,
- interpretacja odpowiedzi HTTP,
- timeout requestu,
- expected_status działa poprawnie,
- expected_body_substring działa poprawnie,
- obliczanie statystyk uptime,
- formatowanie błędów API,
- autoryzacja API key,
- logika incidentów: down tworzy incident, kolejne down zwiększa licznik, up zamyka incident.
Do testów HTTP użyj httptest.Server.
Przykład scenariuszy:
1. Serwer testowy zwraca 200 i body "ok" -> check success.
2. Serwer testowy zwraca 500 -> check failed.
3. Serwer testowy zwraca 200, ale bez oczekiwanego tekstu -> check failed.
4. Serwer testowy śpi dłużej niż timeout -> check failed.
13.2. Testy integracyjne
Minimum:
- utworzenie monitora przez API,
- pobranie monitora przez API,
- ręczne uruchomienie sprawdzenia,
- zapis wyniku w bazie,
- pobranie historii sprawdzeń,
- usunięcie monitora.
Test integracyjny może uruchamiać aplikację z tymczasową bazą SQLite.
13.3. Pokrycie testami
Ustaw cel:
minimum 60% coverage dla internal/
Komenda:
go test ./... -cover
14. Obsługa zamykania aplikacji
Aplikacja musi poprawnie reagować na SIGINT i SIGTERM.
Po naciśnięciu Ctrl+C:
- serwer HTTP przestaje przyjmować nowe requesty,
- scheduler przestaje dodawać nowe zadania,
- workerzy kończą aktualnie wykonywane zadania albo przerywają je przez context timeout,
- połączenie z bazą danych zostaje zamknięte,
- aplikacja kończy się bez panic.
To wymaganie jest ważne, bo nauczy Cię poprawnego użycia context.Context.
15. Docker i uruchamianie
Dodaj Dockerfile.
Wymagania:
- build aplikacji w osobnym etapie,
- finalny obraz powinien zawierać tylko binarkę i potrzebne pliki migracji,
- aplikacja powinna uruchamiać się przez komendę watchdog server.
Dodaj docker-compose.yml dla lokalnego uruchomienia.
Wersja z SQLite może wyglądać koncepcyjnie tak:
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
16. Makefile
Dodaj podstawowe komendy:
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
17. README
README powinno zawierać:
- opis projektu,
- wymagania systemowe,
- instrukcję uruchomienia lokalnego,
- instrukcję uruchomienia przez Docker,
- przykłady użycia API przez curl,
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
18. Etapy realizacji
Etap 1 — szkielet projektu
Zakres:
- utworzenie modułu Go,
- struktura katalogów,
- konfiguracja,
- logger,
- prosty endpoint GET /healthz,
- Makefile.
Kryterium ukończenia:
go run ./cmd/watchdog server
uruchamia serwer, a:
curl http://localhost:8080/healthz
zwraca:
{"status":"ok"}
Etap 2 — baza danych i monitory
Zakres:
- migracje,
- tabela monitors,
- repository dla monitorów,
- endpointy CRUD monitorów,
- walidacja,
- testy walidacji.
Kryterium ukończenia:
Można utworzyć, pobrać, zaktualizować, wyłączyć i usunąć monitor przez REST API.
Etap 3 — checker HTTP
Zakres:
- implementacja wykonywania pojedynczego sprawdzenia,
- timeout przez context,
- obsługa expected_status,
- obsługa expected_body_substring,
- zapis wyniku do bazy,
- testy przez httptest.Server.
Kryterium ukończenia:
Ręczne wywołanie POST /api/v1/monitors/{id}/checks dodaje zadanie, wykonuje request i zapisuje wynik.
Etap 4 — worker pool
Zakres:
- kolejka zadań,
- workerzy,
- obsługa pełnej kolejki,
- graceful shutdown.
Kryterium ukończenia:
Ręczne sprawdzenia nie wykonują się bezpośrednio w handlerze HTTP, tylko przez worker pool.
Etap 5 — scheduler
Zakres:
- cykliczne odczytywanie aktywnych monitorów,
- planowanie sprawdzeń według interwałów,
- brak równoległych sprawdzeń tego samego monitora,
- dynamiczne reagowanie na enable/disable/update/delete.
Kryterium ukończenia:
Po dodaniu aktywnego monitora system sam wykonuje sprawdzenia zgodnie z interwałem.
Etap 6 — status i statystyki
Zakres:
- endpoint statusu monitora,
- endpoint statystyk,
- obliczanie uptime,
- historia sprawdzeń.
Kryterium ukończenia:
Można sprawdzić aktualny status monitora i procent dostępności z wybranego zakresu.
Etap 7 — incidenty i webhooki
Zakres:
- tabela incidents,
- wykrywanie przejść up/down,
- wysyłanie webhooka,
- testy logiki incidentów.
Kryterium ukończenia:
System wysyła webhook tylko wtedy, gdy monitor zmienia stan.
Etap 8 — CLI
Zakres:
- komendy monitor add/list/get/delete/check/status/checks,
- konfiguracja API URL i API key,
- czytelne komunikaty błędów.
Kryterium ukończenia:
Da się korzystać z aplikacji bez pisania curl-i.
Etap 9 — Docker, dokumentacja i porządki
Zakres:
- Dockerfile,
- docker-compose.yml,
- README,
- przykłady użycia,
- testy końcowe,
- go vet,
- coverage.
Kryterium ukończenia:
Nowa osoba może sklonować repozytorium, uruchomić projekt i dodać pierwszy monitor na podstawie README.
19. Definicja ukończenia projektu
Projekt możesz uznać za skończony, gdy spełnia wszystkie poniższe warunki:
1. Aplikacja uruchamia się lokalnie jedną komendą.
2. Można dodać monitor HTTP.
3. Monitor jest automatycznie sprawdzany zgodnie z interwałem.
4. Wyniki sprawdzeń są zapisywane w bazie.
5. Można pobrać historię sprawdzeń przez API.
6. Można pobrać aktualny status monitora.
7. Można pobrać statystyki uptime.
8. Ręczne sprawdzenie działa przez kolejkę i worker pool.
9. Aplikacja obsługuje graceful shutdown.
10. API wymaga klucza X-API-Key.
11. CLI obsługuje podstawowe operacje.
12. Testy przechodzą przez go test ./...
13. README opisuje uruchomienie i przykłady użycia.
14. Projekt można uruchomić przez Docker.
20. Dodatkowe wyzwania, gdy podstawowa wersja będzie gotowa
Po ukończeniu głównej wersji możesz rozszerzyć projekt o:
- prosty panel HTML z html/template,
- eksport wyników do CSV,
- Prometheus metrics pod /metrics,
- endpoint /readyz sprawdzający połączenie z bazą,
- retry policy, np. 3 próby przed oznaczeniem jako down,
- monitorowanie TCP, nie tylko HTTP,
- rate limiting API,
- role użytkowników,
- obsługę wielu kanałów powiadomień: Slack, Discord, e-mail,
- OpenAPI/Swagger,
- benchmarki checkerów,
- property-based tests dla walidacji,
- osobną komendę import/export konfiguracji monitorów.
21. Czego konkretnie się nauczysz
Ten projekt powinien przećwiczyć najważniejsze praktyczne elementy Go:
- projektowanie większej aplikacji w Go,
- czytelny podział na pakiety,
- REST API bez ciężkiego frameworka,
- obsługa JSON,
- praca z bazą przez database/sql,
- migracje SQL,
- context cancellation,
- timeouty,
- goroutines,
- channels,
- worker pool,
- graceful shutdown,
- testy jednostkowe i integracyjne,
- httptest,
- Docker,
- CLI,
- konfiguracja przez zmienne środowiskowe,
- logowanie przez slog,
- obsługa błędów w spójnym formacie.
Najlepsza kolejność: najpierw REST API i baza, potem pojedynczy checker, potem worker pool, potem scheduler. Dzięki temu projekt będzie rósł naturalnie, zamiast od razu zamienić się w trudny do debugowania system współbieżny.
