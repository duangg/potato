#!/bin/sh

PORT=$1
SERVER=$2
CMD=$3
HOST=$4


ssh -tt -p ${PORT} ${SERVER} <<EOF
    #nohup sudo sh /data/backup/xbackup.sh full_backup bkuser bkuser 3306 /data/backup ${PORT} /tmp/mysql.sock /etc/my.cnf &
	${CMD}
	exit 0
EOF
