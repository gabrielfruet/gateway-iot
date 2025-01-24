#!/bin/sh

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

python "$SCRIPT_DIR/main.py" arcondicionado-1 &
python "$SCRIPT_DIR/main.py" arcondicionado-2 & 
python "$SCRIPT_DIR/main.py" arcondicionado-3 &
