# moodle-notifications
a simple moodle grades notifyer that stores all grade changes and reports about important changes

## How to use
1. write login and password or write a token to `moodle-credentials.json`
2. write telegram bot key and telegram chat ID to `telegram-credentials.json`
3. run `go run main.go`

## Tech stack
- **Golang**
- [gjson](https://github.com/tidwall/gjson): tool for JSON parsing
- [diff](https://github.com/r3labs/diff): tool for comparison between structs
- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api): used as an interface to send messages in Telegram
- Docker

## Telegram message example
```
moodle-bot, [9 May 2023 09:30:35]
[S23] English for academic purposes II / Иностранный язык:

Title:  "Timely submission  total"
Grade:  "D"  ->  "C"
Persentage:  "0.00 %"  ->  "50.00 %"

moodle-bot, [9 May 2023 09:30:35]
[S23] Analytical Geometry and Linear Algebra II / Аналитическая геометрия и линейная алгебра II:

Title:  "Link to Assignment activity Joint Assignment 02"
Grade:  ""  ->  "8"
Persentage:  ""  ->  "100 %"


Title:  "Link to Assignment activity Joint Assignment 03"
Grade:  ""  ->  "4"
Persentage:  ""  ->  "100 %"
```
