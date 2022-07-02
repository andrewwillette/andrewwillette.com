#!/usr/bin/env bash

unameOut="$(uname)"

if [ "$unameOut" == "Darwin" ]; then
    for process in $(ps|grep willette_api|awk '! /grep/ {print $1}'); do
        kill "$process"
    done
    for process in $(ps|grep node|awk '!/grep/{print $1}'); do
        kill "$process"
    done
elif [ "$unameOut" == "Linux" ]; then
    for process in $(ps -aux|grep node|awk '! /grep/{print $2}'); do
        kill "$process"
    done
    for process in $(ps -aux|grep willette_api|awk '! /grep/{print $2}'); do
        kill "$process"
    done
fi

