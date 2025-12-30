`configs/` — каталог готовых конфигов и “kitchen sink” шаблона для Reticulum (Go port).

Формат такой же как у Python: это директория, внутри которой лежит файл `config`.
Все утилиты принимают `-config <dir>` (то есть путь именно к директории).

## Быстрый старт

- Запустить один инстанс: `go run ./cmd/rnsd -config ./configs/testing/single_shared_tcp`
- Посмотреть статус: `go run ./cmd/rnstatus -a -config ./configs/testing/single_shared_tcp`

## Структура

- `configs/kitchen_sink/` — “всё в одном” (все ключи/интерфейсы как справочник, в основном `enabled = no`).
- `configs/testing/` — маленькие, детерминированные конфиги, которые удобно использовать в `tests/integration/*.sh`.

## Про несколько инстансов

Для одновременного запуска нескольких `rnsd`:
- используй разные `instance_name`
- и (желательно) `shared_instance_type = tcp` + уникальные `shared_instance_port` / `instance_control_port`
- также разведи порты интерфейсов (UDP/TCP) чтобы не было конфликтов.

Готовые пары:
- `configs/testing/two_nodes_udp/node_a` + `configs/testing/two_nodes_udp/node_b`
- `configs/testing/two_nodes_tcp/server` + `configs/testing/two_nodes_tcp/client`

