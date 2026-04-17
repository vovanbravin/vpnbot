#  VPN Shop Bot

Telegram-бот для автоматической продажи VPN-доступа.

## Быстрый старт
### 1. Клонируйте репозиторий

```bash
git clone https://github.com/vovanbravin/vpnbot.git /home/telegram-bot
cd /home/telegram-bot
```

### 2. Настройте .env
```text
BOT_TOKEN=ваш_токен_от_BotFather
MONGO_URL=mongodb://root:пароль@mongodb:27017
MONGO_DB=vpnbot
```

### 3. Запустите
```bash
docker compose up -d
```
