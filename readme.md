Регистрация curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"имя пользователя", "password":"пароль"}'

Получить токен
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"имя пользователя","password":"пароль"}'

Использовать токен для доступа:
Заменить YOUR_TOKEN на полученный токен

для отслеживания статуса задачи по WebSocket установить wscat и использовать:
wscat -c "ws://localhost:8080/ws" \
  -H "Authorization: Bearer YOUR_TOKEN"

Все задачи: curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks

Добавить задачу:
curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"текст задачи"}'

Прочитать задачу по id:
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/{id}

Обновить задачу:
curl -X PUT http://localhost:8080/tasks/{id} -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"Новый текст"}'

Удалить задачу:
curl -X DELETE -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/{id}



//проверить кэширование:
# 1. Очистка
curl -X POST http://localhost:8080/admin/cache/clear \
  -H "Authorization: Bearer YOUR_TOKEN"

# 2. статистика
curl -X GET http://localhost:8080/admin/cache/stats \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. первый запрос
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_TOKEN"

# 4. статистика
curl -X GET http://localhost:8080/admin/cache/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
# Должно получиться: misses=1, sets=0 или >0

# 5. еще один запрос
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_TOKEN"

# 6. статистика
curl -X GET http://localhost:8080/admin/cache/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
# Должно получиться: hits=1, misses=1