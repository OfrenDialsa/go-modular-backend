#!/bin/sh

#Export System ENV to File because app can't read from system env with (.)dot in name
env > env/.env
date +'%Y-%m-%d/%H:%M' > timerun.txt

#Run Command Specified in Dockerfile
$@