FROM php
RUN apt-get update && apt upgrade -y
COPY . /usr/src/device_service
WORKDIR /usr/src/device_service
EXPOSE 3000
CMD ["php","-S","0.0.0.0:3000"]