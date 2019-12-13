#!/bin/sh

printf "\n**** Begin Environment Variable Dump ****\n\n"
printenv | sort
printf "\n**** End Environment Variable Dump ****\n\n"

./jhuda-docker-user-service serve