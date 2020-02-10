# HandWitch
![](https://github.com/wolf1996/HandWitch/workflows/ci/badge.svg)

Бот для запросов к json api через интерфейс телеграм и формирования ответа в человекочитаемом виде.

## Запуск
Для запуска бота используйте команду serve.
```bash
./HandWitch serve --token=your_telegram_token --config=config.json
```
Более подробную справку по cli можно получить использовав флаг *--help* для каждой команды.

```
HandWitch --help
handwitch helps you to handle http request without frontend

Usage:
  handwitch [command]

Available Commands:
  help        Help about any command
  serve       Starts bot

Flags:
      --config string   configuration path file
  -h, --help            help for handwitch
      --log string      log level [info|warn|debug] (default "info")
      --path string     descriptions file path

Use "handwitch [command] --help" for more information about a command.
```

## Конфигурация 

``` json
{
	"log_level": "Debug", // debug info warning debug panic fatal
	"path": "./descriptions.yaml", // путь до файла с описанием http запросов
	"telegram": { // описание параметров связанных с телеграм
		"white_list": "./whitelist.json", // список логинов пользователей с которыми можно общаться 
		"formatting": "HTML" // разметка 
	}
}
```

Список доступных значений для *log_level* подробней и с описаниями можно посмотреть у [logrus](https://github.com/sirupsen/logrus). 

*formating* - для ответов пользователю можно использовать форматирование текста в формате Markdown ([есть проблема](https://github.com/wolf1996/HandWitch/issues/12)) и HTML. Подробнее про формат можно прочитать в [документации telegram](https://core.telegram.org/bots/api#formatting-options)


*path* - содержит путь до файла, в котором хранится описание запросов формат описания будет приведён ниже.


*white_list* - описание списка пользователей с которыми боту разрешено общаться

``` json
{
    "users": ["YourTelegramName"]
}
```

## Описание запросов

запросы могут быть описны на yaml на yaml по следующему шаблону.

```yaml
urlname:
  url_template: {{url_template}}
  parameters:
    paramname:
      help: помощь параметра
      name: paramname
      destination: URL|query
      type: string|integer
      optional: true
      default_value: defaultvalue
  body: "
    template description of responce body
  "
  url_name: urlname
  help: "help text this hand"
```
где для описания шаблонов *url_template* и *body* используется [шаблонизатор golang](https://golang.org/pkg/text/template/) 

В *url_template* параметры передаются как map напрямую. Т.е. 
```
url_template: http://localhost:8080/{{.string_param}}/{{.int_param}}
```
подставляет *string_param* и *int_param* полученные от пользователя и описанных в *parameters*.

в шаблон же подставляется расширенная структура:
```
{
  "responce": ответсервера,
  "meta": {
    "url":    urlзапроса,
    "params": полученныепараметры,
  }
}
```

пример доступа к параметрам при формировании ответа 
```
Запрошенный url: {{.meta.url}}
Значение paramname: {{.meta.params.paramname}}
Некоторое поле (someresponcefield) в ответе {{.responce.someresponcefield}}
```
## Пользовательская функциональность 

### Начало работы 
Рассмотрим запросы описанные в примере программы. Все конфигурации можно посмотреть в [примере](https://github.com/wolf1996/HandWitch/tree/master/example).

Для начала работы с запросом надо ввести команду.
```
/process example
```
где example - имя запроса из описания. 

После запуска исполнения запроса будет предложено выбрать различные параметры для ввода из столбца кнопок слева. По каждому параметру можно получить справку нажав соответветствующую кнопку в правом столбце.
![Первая клавиатура](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/first_keyboard.png)

Справка для каждого параметра будет выглядеть так:
![Справка по параметру](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/param_help.png)

Для каждого запроса можно получить получить полную справку, или отменить запрос.

![Управление запросом](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/help_keyboard.png)


Справка по запросу включает в себя краткое описание запроса и справку по каждому из его параметров.

![Справка по запросу](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/request_help.png)

Если для запроса не заданы нужные параметры кнопка исполнения запроса бдует скрыта. Пропущенные параметры, необходимые для исполнения запроса будут перечислены как *Missed Params* в сообщении о статусе текущего запроса.

![Статус текущего запроса ](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/state_message.png)

 После того как значения для этих параметров будут установлены, запрос может быть отправлен на исполнение.
![Исполнение запроса](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/ok_params.png)

В результате ответ сервера в json будет форматирован всоответствии с шаблоном. (В данном случае в синтаксисе HTML)
```HTML
  <b>URL requested</b> :\n {{.meta.url}}
  \n
  <b>params</b> :\n
  {{ range $key, $value := .meta.params }} <b>{{ $key }}</b> : {{ $value }} \n {{ end }}
  \n
  <b>responce</b> :\n
  {{ range $key, $value := .responce }} <b>{{ $key }}</b> : {{ $value }} \n {{ end }}
```
![Результат](https://raw.githubusercontent.com/wolf1996/HandWitch/media/pictures/responce.png)
