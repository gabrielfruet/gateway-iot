#!/bin/env bash

#Entre no container sudo docker exec -it devices-iot /bin/bash 
#Rode ./device.sh "arconditioner-1" 5123
#Rode ./device.sh "car_loc-1" 5124
#Rode ./device.sh "dor_lock-1" 5125

# Verifica se os argumentos foram fornecidos
if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <device_name> <port>"
    exit 1
fi

device_name="$1"
port="$2"
hostname="$(hostname -i | cut -d' ' -f1)"  # Obtém o IP do host

# Executa o script Python
python main.py "$device_name" "$hostname" "$port"