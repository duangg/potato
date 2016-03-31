#!/bin/sh

MYSQL_USER=$2
MYSQL_PASSWORD=$3
MYSQL_PORT=$4
MYSQL_HOST=$5
MYSQL_SOCKET=$6
MYSQL_MY=$7
#MYSQL_POSITION=$8
MYSQL_DBNAME=$8
MYSQL_TABLENAME=$9

CHECK_SLAVE_IO="Slave_IO_Running"
CHECK_SLAVE_SQL="Slave_SQL_Running"
CHECK_MASTER_HOST="Master_Host"


isslave(){
	MYSQL_SLAVE=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "show slave status\G") 
    if [ "${MYSQL_SLAVE}" = "" ];then
      echo "0"
    else
      echo "1"
    fi
}

getslavemaster(){
	MYSQL_SLAVE=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "show slave status\G") 
	# MYSQL_SLAVE=$(cat /home/carl/mydoc/dbtest.log)
    MASTERHOST=`echo "${MYSQL_SLAVE}" | grep "${CHECK_MASTER_HOST}" | awk ' {print $2}'`
    echo "${MASTERHOST}"
}

checkslavestatus(){
    MYSQL_SLAVE=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "show slave status\G") 
    # MYSQL_SLAVE=$(cat /home/carl/mydoc/dbtest.log)
    IO_ENV=`echo "${MYSQL_SLAVE}" | grep "${CHECK_SLAVE_IO}" | awk ' {print $2}'` 
    #echo $IO_ENV
    SQL_ENV=`echo "${MYSQL_SLAVE}" | grep "${CHECK_SLAVE_SQL}" | awk '{print $2}'` 
    #echo $SQL_ENV
    if [ "${IO_ENV}" = "Yes" -a "${SQL_ENV}" = "Yes" ];then
      echo "1"
    else
      echo "0"
    fi
}

getmasterslave(){
	mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "select host from information_schema.processlist where command=\"Binlog Dump\";"
}

getdatabase(){
	MYSQL_DATABASES=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "show databases;") 
	echo "${MYSQL_DATABASES}" | sed -n "2,+100p"
}

getdbtable(){
	MYSQL_DATABASES=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "show tables from ${MYSQL_DBNAME};") 
	echo "${MYSQL_DATABASES}" | sed -n "2,+100p"
}

getdbtabledesc(){
	MYSQL_DATABASES=$(mysql --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" -e "use ${MYSQL_DBNAME};select COLUMN_NAME,COLUMN_TYPE,IS_NULLABLE,COLUMN_KEY,COLUMN_DEFAULT,COLUMN_COMMENT from information_schema.columns where table_name = '${MYSQL_TABLENAME}';") 
	echo "${MYSQL_DATABASES}" | sed -n "1,+100p"
	# echo "${MYSQL_DATABASES}"
}

case $1 in
	isslave)
		isslave
		;;
	checkslavestatus)
		checkslavestatus
		;;
	getmasterslave)
		getmasterslave
		;;
	getslavemaster)
		getslavemaster
		;;
	getdatabase)
		getdatabase
		;;
	getdbtable)
		getdbtable
		;;
	getdbtabledesc)
		getdbtabledesc
		;;
	*)
esac
