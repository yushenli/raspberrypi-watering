#!/bin/bash
tmux new-session -d -s rpi -n rpi
sleep 1
tmux send-keys -t rpi:rpi "cd /home/admin" Enter
tmux send-keys -t rpi:rpi "./rpi_run.sh" Enter
tmux attach -t rpi:rpi
