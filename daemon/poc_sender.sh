#!/usr/bin/zsh

while
read -rsk1 c && echo "$c" > /dev/ttyACM0
do true; done
