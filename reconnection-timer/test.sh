#!/bin/bash

set -e


read -e -p 'Input name of service: ' -i 'ohpserver-ssh' SERNAME

echo "Check reconn status by typing systemctl status reconn-$SERNAME"
