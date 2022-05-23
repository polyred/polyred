# Copyright 2022 The Polyred Authors. All rights reserved.
# Use of this source code is governed by a GPLv3 license that
# can be found in the LICENSE file.

FROM golang:1.18
RUN apt-get update && apt-get install -y \
      xvfb libx11-dev libegl1-mesa-dev libgles2-mesa-dev \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
WORKDIR /polyred
COPY . .
CMD [ "sh", "-c", "./test.sh" ]