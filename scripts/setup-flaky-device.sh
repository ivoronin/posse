#!/bin/sh
set -e

DEV=flaky0
DISK=/dev/sdb
RDBLK=0
WRBLK=1

dmsetup remove $DEV 2> /dev/null < /dev/null || true
dmsetup create $DEV <<EOF
#OFF   COUNT MOD    DISK  OFF    UP DOWN NARG FEATURES
$RDBLK 1     flakey $DISK $RDBLK 5  1
$WRBLK 1     flakey $DISK $WRBLK 3  2    5    corrupt_bio_byte 1 w 255 0
EOF
dmsetup table $DEV
