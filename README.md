# Task
Вы решили создать конкурента Amazon S3 и знаете как сделать лучший сервис хранения файлов.
На сервер A по REST присылают файл, его надо разрезать на 6 примерно равных частей и сохранить на серверах хранения Bn (n ≥ 6).
При REST-запросе на сервер A нужно достать куски с серверов Bn склеить их и отдать файл.

Условия:
1. Один сервер для REST запросов
2. Несколько серверов для хранения кусков файлов
3. Файлы могут достигать размера 10 GiB

Ограничения:
1. Реализовать тестовый модуль для сервиса, который обеспечит проверку его функционала, продемонстрирует загрузку и чтение файла.
2. Сервера для хранения могут добавляться в систему в любой момент, но не могут удаляться из системы.
3. Предусмотреть равномерное заполнение серверов хранения.
4. Необходимо учесть различные сценарии, например, пользователь ушел во время загрузки.
5. Сервера хранения должны быть отдельными приложениями. Протокол общения REST сервера с серверами хранения нужно выбрать самостоятельно.
6. Написать docker-compose для поднятия сервиса.
7. Код разместить в Github.

Преимуществом будет:
1. Тестируемый код, речь про unit тесты, при этом сами тесты писать не нужно. В первую очередь будет учитываться подход к решению задачи, однако и на качество кода мы тоже смотрим. 
2. Этим тестовым заданием мы хотим понять образ вашего мышления и умение найти подход к решению задач.

# How to run
1. Clone the repo.
2. Build the image `make build-image`
3. Start 3 store server and 1 front rest server `make up`
4. Now you can use `TestFrontServer` to check that files are uploaded and received successfully.

# Notes
1. REST-service has to have 2 endpoints: uploadFile(fileID, fileContent) and getFile(fileID).
2. In order to not read the whole file to memory (because it can be very large), let's add one more parameter to uploadFile endpoint - fileSize. We use this parameter to split file into chunk in streaming mode and put parts to corresponding server.
3. Front rest server uses badger db to store information about files and its parts, on which server each part is stored.
4. Front rest server gathers statistics from store server once in 10s and use this information to choose least loaded store server. It's not very online, but in big load maybe sufficient.
5. Situation when client cancels file uploading in the middle of the process is not correctly handled now. Currently already uploaded file parts will keep place on store servers forever. I don't have time to implement it, but can explain approach if needed.
6. Current check for existing file is not very consistent. Can be solved by storing more information about files in db. For example - we can store file id before reading any data from request.
7. I made my best to implement features, but current project structure is not ideal. Could explain what I would do if I have more time.
8. I decided to choose http3 protocol over QUIC to achieve ease of development (looks like ordinary webserver) and speed of connection and data transmission.