#!/bin/bash

for file in $(ls *.go) ; do 
  echo -e "${file}の内容です"
  echo '```golang'
  cat $file 
  echo -e '```\n'
done
