package main

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
)

type userProvider interface {
	FromHeaders(headers HeaderProvider) (*User, error)
}

func httpUserService(svc userProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		user, err := svc.FromHeaders(r.Header)
		if err != nil {
			if _, ok := errors.Cause(err).(ErrorBadInput); ok {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("Content-Type", "application/json;charset=utf-8")
		err = user.Serialize(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error encoding JSON response %v", err)
		}
	})
}
