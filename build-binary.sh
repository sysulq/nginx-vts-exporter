#!/bin/sh
random=`awk -v min=10000 -v max=99999 'BEGIN{srand(); print int(min+rand()*(max-min+1))}'`
tag_name="build_$random"
bin_diretory=`pwd`/bin

echo "[Step 1] - Building binary inside image"
docker build --tag=$tag_name \
             --file=Dockerfile-build . 

echo "[Step 2] - Copying the binary from docker image"
docker run --rm --volume=$bin_diretory:/output $tag_name cp /build/nginx-vts-exporter /output/nginx-vts-exporter >/dev/null 2>&1

echo "[Step 3] - Tranfering ownership to current user"
sudo chown -R `whoami`:`whoami` $bin_diretory >/dev/null 2>&1

echo "[Step 4] - Cleaning the tmp build images"
docker rmi -f $tag_name >/dev/null 2>&1

echo "[Success] - Build complete !!"
