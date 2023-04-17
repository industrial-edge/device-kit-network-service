#!/bin/bash

chmod 711 /usr/bin/onboardservice
chmod 640 /lib/systemd/system/dm-onboard.service
systemctl daemon-reload
systemctl enable dm-onboard.service
systemctl restart dm-onboard.service
