# Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
# Use of this source code is governed by a GPLv3 license that
# can be found in the LICENSE file.

all: build test
build:
	sh build.sh
test:
	sh test.sh