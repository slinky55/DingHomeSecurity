#!/bin/bash

ffmpeg -framerate 1 -pattern_type glob -i '*.jpg' -c:v libx264 -r 1 -pix_fmt yuv420p history.mp4 -y