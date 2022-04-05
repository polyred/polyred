# Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
# Use of this source code is governed by a GPLv3 license that
# can be found in the LICENSE file.

# require apt install xvfb
Xvfb :99 -screen 0 1024x768x24 > /dev/null 2>&1 &
export DISPLAY=:99.0

go test -v -covermode=atomic ./internal/gpu