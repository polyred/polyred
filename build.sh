# Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
# Use of this source code is governed by a GPLv3 license that
# can be found in the LICENSE file.

docker build -t poly.red/gpu .
docker run --rm --name gpu poly.red/gpu
docker rmi poly.red/gpu