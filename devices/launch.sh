#!/bin/sh

python "./main.py" arconditioner-1 "$(hostname -i)" 4999 &
python "./main.py" arconditioner-2 "$(hostname -i)" 4998 & 
python "./main.py" arconditioner-3 "$(hostname -i)" 4997 &
