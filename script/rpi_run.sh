#!/bin/sh

cd /home/admin/go/src/github.com/yushenli/raspberrypi-watering
go build -mod=mod .

cd /home/admin
/home/admin/go/src/github.com/yushenli/raspberrypi-watering/raspberrypi-watering \
    --zone-config-filename=/home/admin/rpi_zone_configs.json \
    --status-filename=/home/admin/rpi_water_status.json \
    --check-interval=15s 2>&1 | \
    tee -a /home/admin/rpi_watering.log
