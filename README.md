# ProxyServer

Http прокси сервер и простой сканер уязвимости на его основе.

## Сборка и запуск

```shell
docker build -t proxy .
docker run -d -p 8080:8080 -t proxy
```
## Пример использования

Запрос с помощью программы curl. Задаем адрес прокси сервера в опции -x:
```shell
curl -x http://127.0.0.1:8080 http://mail.ru
```
