package injector

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var addInitContainerPatch string

type admitFunc func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// StartServer - Starts Server
func StartServer(patch, port, urlPath string) {
	addInitContainerPatch = patch
	flag.Parse()

	http.HandleFunc(urlPath, serveMutatePods)
	server := &http.Server{
		Addr: port,
		// Validating cert from client
		// TLSConfig: configTLS(config, getClient()),
	}
	glog.Infof("Starting server at %s", server.Addr)
	var err error
	if os.Getenv("SSL_CRT_FILE_NAME") != "" && os.Getenv("SSL_KEY_FILE_NAME") != "" {
		// Starting in HTTPS mode
		err = server.ListenAndServeTLS(os.Getenv("SSL_CRT_FILE_NAME"), os.Getenv("SSL_KEY_FILE_NAME"))
	} else {
		// LOCAL DEV SERVER : Starting in HTTP mode
		err = server.ListenAndServe()
	}
	if err != nil {
		glog.Errorf("Server Start Failed : %v", err)
	}
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	w.Header().Set("Content-Type", "application/json")
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	glog.V(2).Info(fmt.Sprintf("handling request: %s", string(body)))
	var reviewResponse *admissionv1.AdmissionResponse
	ar := admissionv1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Error(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = admit(ar)
	}
	response := admissionv1.AdmissionReview{}
	response.APIVersion = "admission.k8s.io/v1"
	response.Kind = "AdmissionReview"
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	resp, err := json.Marshal(response)
	if err != nil {
		glog.Error(err)
	}
	glog.V(2).Info(fmt.Sprintf("sending response: %s", string(resp)))
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func serveMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, mutatePods)
}

func mutatePods(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	glog.V(2).Info("mutating pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		glog.Errorf("expect resource to be %s", podResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		glog.Error(err)
		return toAdmissionResponse(err)
	}
	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true
	for k, v := range pod.Annotations {
		if k == "inject-init-container" && strings.ToLower(v) == "true" {
			reviewResponse.Patch = []byte(addInitContainerPatch)
			pt := admissionv1.PatchTypeJSONPatch
			reviewResponse.PatchType = &pt
			return &reviewResponse
		}
	}
	return &reviewResponse
}

func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
