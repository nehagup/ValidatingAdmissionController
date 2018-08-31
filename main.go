package main

import (
	"net/http"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"encoding/json"
	"k8s.io/api/admission/v1beta1"
	"github.com/openshift/api/route/v1"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	CertFile string
	KeyFile  string
}

func main() {

	fmt.Println("Server is going to start at port 8080, This is version 5")

	router := httprouter.New()
	router.POST("/", ServePods)
	router.NotFound = http.HandlerFunc(CustomNotFoundHandler)

	log.Fatal(http.ListenAndServeTLS(":8080","/certificate/tls.crt", "/certificate/tls.key",router))

}


func ServePods(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var t v1beta1.AdmissionReview
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	route := v1.Route{}
	if err := json.Unmarshal(t.Request.Object.Raw, &route); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}
	api := clientset.CoreV1()

	admissionResponse := v1beta1.AdmissionResponse{Allowed:true}

	namespaces, err := api.Namespaces().Get(t.Request.Namespace, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	project_whitelist := namespaces.Labels["router.cern.ch/technical-network-allowed"]


	access := route.Annotations["router.cern.ch/technical-network-access"]
	if access == "true" || access == "True" || access == "TRUE" {
		// route is requesting access to TN
		visiblity := route.Annotations["router.cern.ch/network-visibility"]

		if project_whitelist == "Internet" {
			// all routes are allowed to be exposed to TN
			admissionResponse.Allowed = true
		} else if project_whitelist == "Intranet" {
			// only routes that are NOT visible on Internet can be exposed to TN
			if visiblity == "Internet" || visiblity == "INTERNET" || visiblity == "internet" {
				admissionResponse.Allowed = false
			} else {
				admissionResponse.Allowed = true
			}
		} else {
			// project is not whitelisted for TN access. No route can use TN
			fmt.Println("Label not detected")
			admissionResponse.Allowed = false
		}
	}

	t = v1beta1.AdmissionReview{
		Response: &admissionResponse,
	}

	data, err := json.Marshal(t)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func CustomNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Oops, You're lost!" , r)
}
