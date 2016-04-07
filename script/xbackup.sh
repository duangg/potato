#!/bin/sh

MYSQL_USER=$2
MYSQL_PASSWORD=$3
MYSQL_PORT=$4
MYSQL_SOCKET=$7
MYSQL_MY=$8
MYSQL_BACKUP_DIR=$5
MYSQL_HOST=$6
DAYcurrentDate=`date "+%Y%m%d"`
currentDate=`date "+%Y%m%d%H"`
DELETE_TIME=`date -d "-7 day" "+%Y%m%d"`
DATA_DAY=`date -d '-1 day' '+%Y-%m-%d'`
DATA_TIME_Yesterday=`date -d '-1 day' '+%Y%m%d'`

full_backup(){
	if [ -d "${MYSQL_BACKUP_DIR}" ];then
		mkdir -p "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/ful"
		mkdir -p "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/inc"
		cd "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/ful"
		for DIR in `pwd`
		do
			if [ -d "$DIR" ];then
				/usr/bin/innobackupex --defaults-file="${MYSQL_MY}" --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port="${MYSQL_PORT}" --socket="${MYSQL_SOCKET}" --no-version-check --use-memory=8GB --rsync --tmpdir=/tmp "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/ful" 2>"${MYSQL_BACKUP_DIR}/log"

				if [ $? -eq "0" ];then
					printf "${DAYcurrentDate} innobackupex full backup MySQL 3306 is ok!\n"
				else
					printf "${DAYcurrentDate} innobackupex full backup MySQL 3306 is NO!\n"
					exit 3
				fi
			else
				printf "\n Error ${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/ful is not exist\n" >> "${MYSQL_BACKUP_DIR}/log"
				exit 2
			fi
		done
	else
		printf "\n Error ${MYSQL_BACKUP_DIR} is not exist\n" >> "${MYSQL_BACKUP_DIR}/log"
		exit 1
	fi
}

inc_backup(){
	if [ -d "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}" ];then
		cd "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}"
		for DIR_dir in `ls -al|grep ful|awk '{print $9}'|egrep -v '\.|^$'`
		do
			if [ -d "${DIR_dir}" ];then
				INNOBACKUPEX_FULL_BACKUP_DIR=`ls -al ful/|awk '{print $9}'|egrep -v '\.|^$' | head -n 1`
				for DIR_DIR_dir in `ls -al|grep inc|awk '{print $9}'|egrep -v '\.|^$'`
				do
					if [ -d "${DIR_DIR_dir}" ];then
						/usr/bin/innobackupex --defaults-file="${MYSQL_MY}" --user="${MYSQL_USER}" --password="${MYSQL_PASSWORD}" --host="${MYSQL_HOST}" --port=${MYSQL_PORT} --socket="${MYSQL_SOCKET}" --no-version-check --use-memory=8GB --rsync --tmpdir=/tmp --incremental --incremental-basedir="${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/ful/${INNOBACKUPEX_FULL_BACKUP_DIR}" "${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/inc/" 2>"${MYSQL_BACKUP_DIR}/log"
						if [ $? -eq "0" ];then
							printf "${currentDate} innobackupex inc backup MySQL 3306 is ok!\n"
						else
							printf "${currentDate} innobackupex inc backup MySQL 3306 is no!\n"
							exit 6
						fi
					else
						printf "\n Error ${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/${DIR_DIR_dir} is not exist\n" >> "${MYSQL_BACKUP_DIR}/log"
						exit 5
					fi
				done
			else
				printf "\n Error ${MYSQL_BACKUP_DIR}/${DAYcurrentDate}/${DIR_dir} is not exist\n" >> "${MYSQL_BACKUP_DIR}/log"
				exit 7
			fi
		done
	else
		printf "\n Error ${MYSQL_BACKUP_DIR}/${DAYcurrentDate} is not exist\n" >> "${MYSQL_BACKUP_DIR}/log"
		exit 4
	fi
}

scp_lftp(){
    echo "" > "${MYSQL_BACKUP_DIR}/compresslog"
	if [ -d "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful/" ];then
        cd "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}"/ful/
        for ful_dir in `ls -al|grep $DATA_DAY|awk '{print $9}'|egrep -v '\.|^$'`
        do
            tar -zcvf "${ful_dir}".tar.gz "${ful_dir}"
			if [ $? -eq "0" ];then
               	printf  "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful/${ful_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
        	else
             	printf "ERROR:failed to tar ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful/${ful_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
            fi
			rm -rf  ${ful_dir}
			if [ $? -eq "0" ];then
			    printf  "succeed to delete ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful/${ful_dir}\n"
	    	else
              	printf "ERROR:failed to remove ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful/${ful_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
            fi
        done
    else
        printf "ERROR:${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/ful is not exist\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
    fi

    if [ -d "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc/" ];then
        cd "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}"/inc/
        for inc_dir in `ls -al|grep $DATA_DAY|awk '{print $9}'|egrep -v '\.|^$'`
        do
            tar -zcvf "${inc_dir}".tar.gz "${inc_dir}"
        	if [ $? -eq "0" ];then
            	printf  "${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc/${inc_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
        	else
               	printf "ERROR:failed to tar ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc/${inc_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
            fi
            rm -rf "${inc_dir}"
            if [ $? -eq "0" ];then
                printf  "succeed to delete ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc/${inc_dir}\n"
	    	else
          		printf "ERROR:failed to remove ${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc/${inc_dir}\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
       		fi
        done
    else
        printf "ERROR:${MYSQL_BACKUP_DIR}/${DATA_TIME_Yesterday}/inc is not exist\n" >> "${MYSQL_BACKUP_DIR}/compresslog"
    fi
}

# To determine parameters
NUMB=0
[ $# -eq "${NUMB}" ] && echo "Ple input [options]" && exit 7

case $1 in
	full_backup)
		full_backup
		;;
	inc_backup)
		inc_backup
		;;
	scp_lftp)
		scp_lftp
		;;
	*)
		echo "Usage: $(basename $0) [OPTION] (full_backup|inc_backup|scp_lftp)"
		exit 8
esac

