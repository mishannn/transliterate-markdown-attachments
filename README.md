# Транслитератор вложений в Markdown

Сделан чтобы массово переименовывать файлы и обновлять ссылки в MD файлах.
Поддерживает откат в случае возникновения ошибок во время конвертации.

## Сборка

```
go build -o transliterate-markdown-attachments
```

## Использование

```
./transliterate-markdown-attachments convert -f /path/to/file
```

## Справка

```
NAME:
   transliterate-markdown-attachments convert

USAGE:
   transliterate-markdown-attachments convert [command options]

OPTIONS:
   --file value, -f value  path to markdown file
   --help, -h              show help
```