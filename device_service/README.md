# Docker Setup

## Running the service

```
make build && make serve
```

# Deployment

# Uploading the PHP service

```
make pack upload serve
```

# Testing Local Setup

```
curl --location --request POST 'http://localhost:3000/device_service' \
--header 'User-Agent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"'
```

# Local Setup

## Installing php

```
 brew install php
```

## Composer installation

```
php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
php -r "if (hash_file('sha384', 'composer-setup.php') === '55ce33d7678c5a611085589f1f3ddf8b3c52d662cd01d4ba75c0ee0459970c2200a51f492d557530c71c15d8dba01eae') { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;"
php composer-setup.php
php -r "unlink('composer-setup.php');"
php composer.phar install
mv composer.phar /usr/local/bin/composer
open -e  ~/.zshrc or ~/.bash_profile
```

Then add below command to ~/.zshrc or ~/.bash_profile :

```
alias composer="php /usr/local/bin/composer"
```

## Installing matomo/device-detector

```
composer require matomo/device-detector
composer install
```

## Updating matomo/device-detector

```
composer update matomo/device-detector
```