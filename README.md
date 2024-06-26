# Gophermart

**Order**

Number - уникальное число
status (NEW, PROCESSING, INVALID, PROCESSED) - нужна ли в базе
UserId
UploadedAt
ProcessedAt
Sum


**User**
id - нужен ли
login
password (hash)
current
withdrawn


**Sessions**

id
userId
refresh_token
expired
FOREIGN KEY(`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE