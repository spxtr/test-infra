package fake

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"

	"k8s.io/test-infra/prow/kube"
)

// shiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
// From http://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type Server struct {
	// Set of namespaces.
	Namespaces map[string]bool
	// All are maps from namespace to list of resources.
	ConfigMaps map[string][]kube.ConfigMap
	Pods       map[string][]kube.Pod
	ProwJobs   map[string][]kube.ProwJob
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := shiftPath(r.URL.Path)
	switch head {
	case "api":
		s.handleAPI(w, r, tail)
	case "apis":
	default:
		http.Error(w, "Unknown request path: "+head, http.StatusNotFound)
	}
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request, path string) {
	head, tail := shiftPath(path)
	if head != "v1" {
		http.Error(w, "Unknown request path: "+head, http.StatusNotFound)
		return
	}
	head, tail = shiftPath(tail)
	if head != "namespaces" {
		http.Error(w, "Unknown request path: "+head, http.StatusNotFound)
		return
	}
	ns, tail := shiftPath(tail)
	if !s.Namespaces[ns] {
		http.Error(w, "Namespace not found: "+ns, http.StatusNotFound)
		return
	}
	head, tail = shiftPath(tail)
	switch head {
	case "pods":
		s.handlePods(w, r, ns, tail)
	case "configmaps":
	default:
		http.Error(w, "Unknown request path: "+head, http.StatusNotFound)
	}
}

func (s *Server) handlePods(w http.ResponseWriter, r *http.Request, ns, path string) {
	head, tail := shiftPath(path)
	if head == "" && r.Method == http.MethodGet {
		// ListPods
		if err := json.NewEncoder(w).Encode(s.Pods[ns]); err != nil {
			panic(err)
		}
	} else if head != "" && tail == "" && r.Method == http.MethodGet {
		// GetPod
		for _, p := range s.Pods[ns] {
			if p.Metadata.Name == head {
				if err := json.NewEncoder(w).Encode(&p); err != nil {
					panic(err)
				}
				return
			}
		}
		http.Error(w, "Pod not found: "+head, http.StatusNotFound)
	}
}
