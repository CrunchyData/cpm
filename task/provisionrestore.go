/*
 Copyright 2015 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package task

import (
	"database/sql"
	"github.com/crunchydata/crunchy-postgresql-manager/admindb"
	"github.com/crunchydata/crunchy-postgresql-manager/cpmserverapi"
	"github.com/crunchydata/crunchy-postgresql-manager/logit"
	"github.com/crunchydata/crunchy-postgresql-manager/types"
)

// ProvisionRestoreJob creates a docker container to orchestrate a restore job
func ProvisionRestoreJob(dbConn *sql.DB, args *TaskRequest) error {

	logit.Info.Println("task.ProvisionRestoreJob called")
	logit.Info.Println("with scheduleid=" + args.ScheduleID)
	logit.Info.Println("with serverid=" + args.ServerID)
	logit.Info.Println("with servername=" + args.ServerName)
	logit.Info.Println("with serverip=" + args.ServerIP)
	logit.Info.Println("with containername=" + args.ContainerName)
	logit.Info.Println("with profilename=" + args.ProfileName)

	params := &cpmserverapi.DockerRunRequest{}
	params.Image = "crunchydata/cpm-restore-job"
	params.ServerID = args.ServerID
	restorecontainername := args.ContainerName + "-restore-job"
	params.ContainerName = restorecontainername
	params.Standalone = "false"

	//get server info
	server, err := admindb.GetServer(dbConn, params.ServerID)
	if err != nil {
		logit.Error.Println(err.Error())
		return err
	}

	schedule, err := GetSchedule(dbConn, args.ScheduleID)
	if err != nil {
		logit.Error.Println(err.Error())
		return err
	}

	//params.PGDataPath = server.PGDataPath + "/" + restorecontainername + "/" + getFormattedDate()

	//get the docker profile settings
	var setting types.Setting
	setting, err = admindb.GetSetting(dbConn, "S-DOCKER-PROFILE-CPU")
	params.CPU = setting.Value
	setting, err = admindb.GetSetting(dbConn, "S-DOCKER-PROFILE-MEM")
	params.MEM = setting.Value

	params.EnvVars = make(map[string]string)

	params.EnvVars["RestoreRemotePath"] = schedule.RestoreRemotePath
	params.EnvVars["RestoreRemoteHost"] = schedule.RestoreRemoteHost
	params.EnvVars["RestoreRemoteUser"] = schedule.RestoreRemoteUser
	params.EnvVars["RestoreDbUser"] = schedule.RestoreDbUser
	params.EnvVars["RestoreDbPass"] = schedule.RestoreDbPass
	params.EnvVars["RestoreSet"] = schedule.RestoreSet
	params.EnvVars["RestoreContainerName"] = args.ContainerName
	params.EnvVars["RestoreScheduleID"] = args.ScheduleID
	params.EnvVars["RestoreServerName"] = server.Name
	params.EnvVars["RestoreServerIP"] = server.IPAddress
	params.EnvVars["RestoreProfileName"] = args.ProfileName

	setting, err = admindb.GetSetting(dbConn, "PG-PORT")
	if err != nil {
		logit.Error.Println(err.Error())
		return err
	}
	params.EnvVars["RestorePGPort"] = setting.Value

	//run the container
	params.CommandPath = "docker-run-restore.sh"
	var response cpmserverapi.DockerRunResponse
	var url = "http://" + server.IPAddress + ":10001"
	response, err = cpmserverapi.DockerRunClient(url, params)
	if err != nil {
		logit.Error.Println(response.Output)
		return err
	}
	logit.Info.Println("docker-run-restore.sh output=" + response.Output)

	return nil
}
