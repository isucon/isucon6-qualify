#!/bin/bash

ansible-playbook -i localhost, ./java_deploy.yml --connection=local -t dev