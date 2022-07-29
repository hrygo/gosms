#!/bin/sh

./make.sh

mv gosmsn.tar.gz ~/gosmsn
rm -f gosmsn

tar zxf ~/gosmsn/gosmsn.tar.gz -C ~/gosmsn

