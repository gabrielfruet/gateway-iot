#!/bin/sh

python "./main.py" arconditioner-1 localhost 4999 &
python "./main.py" arconditioner-2 localhost 4998 & 
python "./main.py" arconditioner-3 localhost 4997 &
