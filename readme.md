Регистрация curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"имя пользователя", "password":"пароль"}'

Получить токен
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"имя пользователя","password":"пароль"}'

Использовать токен для доступа:
Заменить YOUR_TOKEN на полученный токен

Все задачи: curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks

Добавить задачу:
curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"текст задачи"}'

Прочитать задачу по id:
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/{id}

Обновить задачу:
curl -X PUT http://localhost:8080/tasks/{id} -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"Новый текст"}'

Удалить задачу:
curl -X DELETE -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/tasks/{id}