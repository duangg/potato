package store

import (
	"time"
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type JobResultStore struct {
	*MongoStore
}

const (
	ColJobResults = "job_results"
)

func NewJobResultStore(mongo *MongoStore) *JobResultStore {
	return &JobResultStore{mongo}
}

func getJobResultCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColJobResults)
}

func (j *JobResultStore) SaveJobResult(serverName, jobId, err string, starttime, endtime time.Time, backuplog string, result, exectype int) error {
	session := j.GetSession()
	defer session.Close()

	var jobresult models.JobResult
	error := getJobResultCol(session).Find(bson.M{"servername": serverName, "job_id": bson.ObjectIdHex(jobId), "starttime": starttime}).One(&jobresult)
	if error != nil {
		jobresult := models.JobResult{
			Id:         bson.NewObjectId(),
			ServerName: serverName,
			JobId:      bson.ObjectIdHex(jobId),
			Status:     result,
			ErrInfo:    err,
			StartTime:  starttime,
			EndTime:    endtime,
			Output:     backuplog,
			Type:       exectype,
		}
		return getJobResultCol(session).Insert(jobresult)
	} else {
		if ((jobresult.Status == models.JobFailed) || (jobresult.Status == models.JobSucceeded)){
			return errors.New("The job is in error or finish status")
		}
		jobresult.EndTime = endtime
		jobresult.Status = result
		jobresult.ErrInfo = err
		jobresult.Output = backuplog
		return getJobResultCol(session).Update(bson.M{"servername": serverName, "job_id": bson.ObjectIdHex(jobId), "starttime": starttime}, jobresult)
	}
}

func (j *JobResultStore) FindJobResultByHostName(name string) ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"servername": name}).All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) FindJobResultByCronId(id string) ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"job_id": bson.ObjectIdHex(id)}).Sort("-starttime").All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) FindLastJobResultByCronId(id string, num int) ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"job_id": bson.ObjectIdHex(id)}).Sort("-starttime").Limit(num).All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) FindPageJobResultByCronId(id string, pagenum, pagesize int) ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"job_id": bson.ObjectIdHex(id)}).Sort("-starttime").Skip(pagenum * pagesize).Limit(pagesize).All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) FindAllBackupErrorResult() ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"status": models.JobFailed, "dismiss": false}).Sort("servername", "-starttime", "type").All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) FindAllBackupRunningResult() ([]models.JobResult, error) {
	session := j.GetSession()
	defer session.Close()
	var jobresults []models.JobResult
	err := getJobResultCol(session).Find(bson.M{"status": models.JobInProgress}).Sort("servername", "-starttime", "type").All(&jobresults)
	return jobresults, err
}

func (j *JobResultStore) DismissBackupErrorResult(id string) error {
	session := j.GetSession()
	defer session.Close()
	var jobresult models.JobResult
	error := getJobResultCol(session).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&jobresult)
	if error != nil {
		return error
	}
	if jobresult.Status != models.JobFailed {
		return errors.New("The job is not error status")
	}
	jobresult.Dismiss = true
	return getJobResultCol(session).Update(bson.M{"_id": bson.ObjectIdHex(id)}, jobresult)
}

func (j *JobResultStore) AbortRunningBackupResult(id, err string) (error, string) {
	session := j.GetSession()
	defer session.Close()
	var jobresult models.JobResult
	error := getJobResultCol(session).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&jobresult)
	if error != nil {
		return error, ""
	}
	//the job is not in running, return error
	if jobresult.Status != models.JobInProgress {
		return errors.New("The job is not running status"), ""
	}
	jobresult.Status = models.JobFailed
	jobresult.ErrInfo = err
	jobresult.EndTime = time.Now()
	return getJobResultCol(session).Update(bson.M{"_id": bson.ObjectIdHex(id)}, jobresult), jobresult.JobId.Hex()
}
