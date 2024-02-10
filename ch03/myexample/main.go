package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"net/http"
)

func main() {
	err := createPod()
	if err != nil {
		panic(err)
	}
}

func createPod() error {
	pod := createPodObject()
	serializer := getJSONSerializer()

	postBody, err := serializePodObject(serializer, pod)
	if err != nil {
		return err
	}
	post, err := createPostRequest(postBody)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(post)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 300 { // ➏
		createdPod, err := deserializePodBody(serializer, body) // ➐
		if err != nil {
			return err
		}
		json, err := json.MarshalIndent(createdPod, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json) // ➑
	} else {
		status, err := deserializeStatusBody(serializer, body) // ➒
		if err != nil {
			return err
		}
		json, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", json) // ➓
	}
	return nil
}

func createPodObject() *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "mus",
			Labels: map[string]string{
				"app.kubernetes.io/component": "my-component",
				"app.kubernetes.io/name":      "a-name",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "runtime",
					Image: "nginx",
				},
			},
		},
	}

	return pod
}

func serializePodObject(serializer runtime.Serializer, pod *v1.Pod) (io.Reader, error) {
	var buf bytes.Buffer
	err := serializer.Encode(pod, &buf)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func createPostRequest(r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest("POST", "http://127.0.0.1:8001/api/v1/namespaces/mus/pods", r)
	if err != nil {
		return nil, err
	}

	req.Header.Add(
		"Accept",
		"application/json",
	)
	req.Header.Add(
		"Content-Type",
		"application/json",
	)

	return req, nil
}

func deserializePodBody(serializer runtime.Serializer, body []byte) (*v1.Pod, error) {

	pod := v1.Pod{}
	_, _, err := serializer.Decode(body, nil, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func deserializeStatusBody(serializer runtime.Serializer, body []byte) (*metav1.Status, error) {
	status := metav1.Status{}
	_, _, err := serializer.Decode(body, nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func getJSONSerializer() runtime.Serializer {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(
		schema.GroupVersion{
			Group:   "",
			Version: "v1",
		},
		&v1.Pod{},
		&metav1.Status{},
	)
	return kjson.NewSerializerWithOptions(
		kjson.SimpleMetaFactory{},
		nil,
		scheme,
		kjson.SerializerOptions{},
	)
}
