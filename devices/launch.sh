#!/bin/env bash

devices=(
    "arconditioner-1" 
    "arconditioner-2" 
    "arconditioner-3" 
    "light-1" 
    "temperature_sensor-2" 
    "door_lock-3"
    "car_loc-1"
)

base_port=5123

pids=()

for i in "${!devices[@]}"; do
    device_name="${devices[$i]}"
    port="$((i + base_port))"
    hostname="$(hostname -i | cut -d' ' -f1)"
    python main.py "$device_name" "$hostname" "$port" &
    pids+=( $! )
done

on_exit() {
    for i in "${!pids[@]}"; do
        kill "${pids[$i]}"
        echo "Killing device ${devices[$i]} of pid ${pids[$i]}"
    done
}

trap on_exit EXIT

for pid in "${pids[@]}"; do
    wait "$pid"
done
