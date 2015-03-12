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

package main

import (
	"bytes"
	"crunchy.com/template"
	"crypto/tls"
	"crypto/x509"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang/glog"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func TestCreate(w rest.ResponseWriter, r *rest.Request) {

	glog.Infoln("here in Test Create")

	kubeURL := os.Getenv("KUBE_URL")
	if kubeURL == "" {
		glog.Errorln("TestCreate: KUBE_URL not set")
		rest.Error(w, "KUBE_URL not set", http.StatusBadRequest)
	}

	podInfo := template.KubePodParams{
		"testnode",
		"0", "0",
		"crunchydata/cpm-node",
		"/opt/cpm/data/pgsql/testnode"}
	err := CreatePod(kubeURL, podInfo)
	if err != nil {
		glog.Infoln(err.Error())
		glog.Errorln("TestCreate:" + err.Error())
	}
	glog.Infoln("no error on create pod")

	response := KubeResponse{}
	response.URL = "here in TestCreate"
	w.WriteHeader(http.StatusOK)
	w.WriteJson(&response)
}

func TestDelete(w rest.ResponseWriter, r *rest.Request) {

	glog.Infoln("here in Test Delete")
	kubeURL := os.Getenv("KUBE_URL")
	if kubeURL == "" {
		glog.Errorln("TestDelete: KUBE_URL not set")
		rest.Error(w, "KUBE_URL not set", http.StatusBadRequest)
		return
	}
	err := DeletePod(kubeURL, "testnode")
	if err != nil {
		glog.Infoln(err.Error())
		glog.Errorln("TestCreate:" + err.Error())
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	glog.Infoln("no error on delete pod")
	response := KubeResponse{}
	response.URL = "here in TestDelete"
	w.WriteHeader(http.StatusOK)
	w.WriteJson(&response)
}

// DeletePod deletes a kube pod that should already exist
// kubeURL  - the URL to kube
// ID - the ID of the Pod we want to delete
// it returns an error is there was a problem
func DeletePod(kubeURL string, ID string) error {
	glog.Infoln("deleting pod " + ID)

	var caFile = "/kubekeys/root.crt"
	var certFile = "/kubekeys/cert.crt"
	var keyFile = "/kubekeys/key.key"

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// DELETE
	var url = kubeURL + "/api/v1beta1/pods/" + ID
	glog.Infoln("url is " + url)
	request, err2 := http.NewRequest("DELETE", url, nil)
	if err2 != nil {
		glog.Errorln(err2.Error())
		return err2
	}

	resp, err := client.Do(request)
	if err != nil {
		glog.Errorln(err.Error())
		return err
	}
	defer resp.Body.Close()

	// Dump response
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		glog.Errorln(err2.Error())
		return nil
	}
	log.Println(string(data))

	return nil
}

// CreatePod creates a new pod using passed in values
// kubeURL - the URL to the kube
// podInfo - the params used to configure the pod
// return an error if anything goes wrong
func CreatePod(kubeURL string, podInfo template.KubePodParams) error {
	var caFile = "/kubekeys/root.crt"
	var certFile = "/kubekeys/cert.crt"
	var keyFile = "/kubekeys/key.key"

	glog.Infoln("creating pod " + podInfo.ID)

	//use a pod template to build the pod definition
	data, err := template.KubeNodePod(podInfo)
	if err != nil {
		glog.Errorln("CreatePod:" + err.Error())
		return err
	}

	glog.Infoln(string(data[:]))

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// POST
	var bodyType = "application/json"
	var url = kubeURL + "/api/v1beta1/pods"
	glog.Infoln("url is " + url)
	resp, err := client.Post(url, bodyType, bytes.NewReader(data))
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}
	defer resp.Body.Close()

	// Dump response
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		glog.Errorln(err2.Error())
		return nil
	}
	log.Println(string(data))

	return nil

}

// CreatePod creates a new pod using passed in values
// kubeURL - the URL to the kube
// podInfo - the params used to configure the pod
// return an error if anything goes wrong
func GetPods(kubeURL string, podInfo template.KubePodParams) error {
	var caFile = "/kubekeys/root.crt"
	var certFile = "/kubekeys/cert.crt"
	var keyFile = "/kubekeys/key.key"

	glog.Infoln("creating pod " + podInfo.ID)

	//use a pod template to build the pod definition
	data, err := template.KubeNodePod(podInfo)
	if err != nil {
		glog.Errorln("CreatePod:" + err.Error())
		return err
	}

	glog.Infoln(string(data[:]))

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// Do GET something
	resp, err := client.Get(kubeURL + "/api/v1beta1/pods")
	if err != nil {
		glog.Errorln(err.Error())
		return nil
	}
	defer resp.Body.Close()

	// Dump response
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		glog.Errorln(err2.Error())
		return nil
	}
	log.Println(string(data))

	return nil
}
