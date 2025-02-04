#!/bin/bash

# NOTE Run this Script as a ROOT user 

# Update the pacakages
echo "updaing pacakages"
apt-get update

# Install docker
echo "Installing docker.io" 
apt-get install docker.io -y

# Enable docker
echo "enable docker-engine"
systemctl enable docker

# Start docker
echo "start docker-engine"
systemctl start docker

# Install docker-compose
echo "installing docker-compose"
apt-get install docker-compose -y

echo "updating /etc/hosts"
echo '127.0.0.1     factors-dev.com' | sudo tee -a /etc/hosts

# Start docker services
echo "start docker serive memsql redis api"
docker-compose up -d memsql redis api

# Install ngrok and expose the service
cd ~
echo "instaiing ngrok .tgz file"
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz


tar xvzf ./ngrok-v3-stable-linux-amd64.tgz -C /usr/bin/

echo "Removing .tgz file"
rm  ngrok-v3-stable-linux-amd64.tgz 

echo "adding ngrok auth token"
ngrok config add-authtoken $NGROK_AUTH

echo "CD .config/ngrok/"
cd .config/ngrok/

echo "upating ngrok.yml"

ngrok_conf_data="
tunnels:
    api_service:
        addr: 8080
        proto: http                                  
    memsql_service:
        addr: 8040
        proto: http"

echo "$ngrok_conf_data" | sudo sed -i '2 r /dev/stdin' "ngrok.yml"


# Creating ngrok.service file
cd ~
echo "Creating ngrok.service file"
cd  /etc/systemd/system/

ngrok_service_data="[Unit]
Description=Ngrok
After=network.service
[Service]
Type=simple
User=root
WorkingDirectory=/root
ExecStart=/usr/bin/ngrok start --all --config=\"/root/.config/ngrok/ngrok.yml\"
Restart=on-failure
[Install]
WantedBy=multi-user.target"

echo "$ngrok_service_data" > ngrok.service

# Enable and start ngrok service
cd ~
ehco "Enable and start ngrok service"
systemctl enable ngrok & systemctl start ngrok
sleep 5

# Install python
echo "installing python "
apt-get install python -y 


public_urls=$(curl -s http://localhost:4040/api/tunnels | python -c "import json, sys; data = json.load(sys.stdin); print('\n'.join([t['public_url'] for t in data['tunnels']]))")

echo "Ngrok URLs:"
echo "$public_urls"






