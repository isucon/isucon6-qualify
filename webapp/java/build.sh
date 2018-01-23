#!/bin/bash

{
  cd /home/isucon/webapp/java/isuda
  gradle -q fatjar
  cp -f build/libs/isuda.jar ../
}

{
  cd /home/isucon/webapp/java/isutar
  gradle -q fatjar
  cp -f build/libs/isutar.jar ../
}