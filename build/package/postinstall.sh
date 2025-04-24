#!/bin/bash

chmod 711 /usr/bin/networkservice
chmod 640 /lib/systemd/system/dm-network.service
systemctl daemon-reload
systemctl enable dm-network.service
systemctl restart dm-network.service
