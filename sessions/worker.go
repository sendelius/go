package sessions

import (
	"log"
	"time"
)

type Worker[T Accessor] struct {
	sessions *Service[T]
}

func newWorker[T Accessor](sessions *Service[T]) *Worker[T] {
	return &Worker[T]{
		sessions: sessions,
	}
}

// Clean отчистка старых сессий каждый час
func (w *Worker[T]) Clean() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		if err := w.sessions.RemoveExpired(); err != nil {
			log.Printf("ошибка очистки сессий: %v", err)
		}
		<-ticker.C
	}
}
