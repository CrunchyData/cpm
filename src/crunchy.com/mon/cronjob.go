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

package mon

import (
	"crunchy.com/admindb"
	"crunchy.com/logit"
	"crunchy.com/util"
	"database/sql"
	"github.com/myinfluxdb/client"
	"strconv"
	"time"
)

type DefaultJob struct {
	request MonRequest
}

//this is the func that implements the cron Job interface
func (t DefaultJob) Run() {
	logit.Info.Println("running Schedule:" + t.request.Schedule.Name)
	RunMonJob(&t.request)
}

func RunMonJob(args *MonRequest) error {
	logit.Info.Println("mon.RunMonJob called")
	logit.Info.Println("with Schedule Name=" + args.Schedule.Name)

	var scheduleTS = time.Now().Unix()

	c, err := client.NewClient(&client.ClientConfig{
		Username: "root",
		Password: "root",
		Database: "cpm",
	})
	if err != nil {
		logit.Error.Println("error in connection to fluxdb " + err.Error())
		return err
	}

	domain, servers, nodes, metrics, err := getData()
	if err != nil {
		logit.Error.Println("error: RunMonJob " + err.Error())
		return err
	}

	var value DBMetric
	var values []DBMetric
	x := 0
	for x = range servers {
		//get connection to server
		logit.Info.Println("collecting for server " + servers[x].Name)
		//collect all the server metrics
		i := 0
		for i = range metrics {
			if metrics[i].ScheduleName == args.Schedule.Name {
				if metrics[i].MetricType == "server" {
					value, err = collectServerMetrics(metrics[i].Name, servers[x].IPAddress)
					if err != nil {
						logit.Error.Println(err.Error())
					} else {
						//add value to influxdb here
						logit.Info.Println(metrics[i].Name + " value " + strconv.FormatFloat(value.Value, 'f', 3, 64))
						series := &client.Series{
							Name:    metrics[i].Name,
							Columns: []string{"value", "server"},
							Points: [][]interface{}{
								{value.Value, servers[x].Name},
							},
						}
						if err = c.WriteSeries([]*client.Series{series}); err != nil {
							logit.Error.Println("error writing to influxdb " + err.Error())
						}
					}

				} else if metrics[i].MetricType == "healthck" {
					logit.Info.Println("healthck server metrics callled on " + args.Schedule.Name)
				}
			}
		}
	}

	var databaseConn *sql.DB
	var id = "postgres"
	var psw = ""
	var database = "postgres"
	y := 0
	for y = range nodes {
		//get connection to database
		logit.Info.Println("collecting for node " + nodes[y].Name)

		if nodes[y].Role == "pgpool" {
			id = "cpmtest"
			psw = "cpmtest"
			database = "cpmtest"
		}

		databaseConn, err = util.GetMonitoringConnection(nodes[y].Name+"."+domain, id, "5432", database, psw)
		if err != nil {
			logit.Error.Println("error in getting connection to " + nodes[y].Name)
		} else {
			//collect all the database metrics
			i := 0
			for i = range metrics {
				if metrics[i].ScheduleName == args.Schedule.Name {
					if metrics[i].MetricType == "database" {
						values, err = collectContainerMetrics(metrics[i].Name, databaseConn)
						if err != nil {
							logit.Error.Println(err.Error())
						} else {
							j := 0
							for j = range values {
								logit.Info.Println(metrics[i].Name + " value " + strconv.FormatFloat(values[j].Value, 'f', 3, 64))
								//add value to influxdb here
								series := &client.Series{
									Name:    metrics[i].Name,
									Columns: []string{"value", "container", "database"},
									Points: [][]interface{}{
										{values[j].Value, nodes[y].Name, values[j].Name},
									},
								}
								if err = c.WriteSeries([]*client.Series{series}); err != nil {
									logit.Error.Println("error writing to influxdb " + err.Error())
								}
							}
						}
					} else if metrics[i].MetricType == "healthck" {
						logit.Info.Println("healthck metric database run on schedule " + args.Schedule.Name)
						//hc1 - database down condition
						hc1(scheduleTS, nodes[y].Name, databaseConn, c)
					}
				}
			}
		}
		databaseConn.Close()
	}

	return err

}

func getData() (string, []admindb.DBServer, []admindb.DBClusterNode, []MonMetric, error) {
	var domain string
	var servers []admindb.DBServer
	var nodes []admindb.DBClusterNode
	var metrics []MonMetric
	var dbConn *sql.DB
	var err error

	dbConn, err = util.GetConnection("clusteradmin")
	if err != nil {
		logit.Error.Println(err.Error())
		return domain, servers, nodes, metrics, err
	}
	defer dbConn.Close()

	admindb.SetConnection(dbConn)

	domain, err = admindb.GetDomain()
	if err != nil {
		logit.Error.Println("error: getData " + err.Error())
		return domain, servers, nodes, metrics, err
	}

	servers, err = admindb.GetAllDBServers()
	if err != nil {
		logit.Error.Println(err.Error())
		return domain, servers, nodes, metrics, err
	}
	nodes, err = admindb.GetAllDBNodes()
	if err != nil {
		logit.Error.Println(err.Error())
		return domain, servers, nodes, metrics, err
	}

	SetConnection(dbConn)
	metrics, err = DBGetMetrics()
	if err != nil {
		logit.Error.Println(err.Error())
		return domain, servers, nodes, metrics, err
	}

	logit.Info.Println("got this many metrics " + strconv.Itoa(len(metrics)))
	return domain, servers, nodes, metrics, err
}
