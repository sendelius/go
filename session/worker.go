package session

import (
	"log"
	"time"
)

type Worker struct {
	sessions *Service
}

func NewWorker(sessions *Service) *Worker {
	return &Worker{sessions: sessions}
}

func (w *Worker) Clean() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		if err := w.sessions.RemoveExpired(); err != nil {
			log.Printf("ошибка очистки сессий: %v", err)
		}
		<-ticker.C
	}
}
