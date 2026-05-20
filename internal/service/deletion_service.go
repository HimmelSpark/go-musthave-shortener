package service

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/HimmelSpark/go-musthave-shortener.git/internal/repository"
)

type DeletionService interface {
	Submit(userID string, shortIDs []string)
	Run(ctx context.Context)
}

type deletionJob struct {
	userID  string
	shortID string
}

type deletionService struct {
	repo          repository.URLRepository
	jobs          chan deletionJob
	bufSize       int
	flushInterval time.Duration
	dropped       atomic.Uint64
}

func NewDeletionService(repo repository.URLRepository, channelBuf, bufSize int, flushInterval time.Duration) DeletionService {
	return &deletionService{
		repo:          repo,
		jobs:          make(chan deletionJob, channelBuf),
		bufSize:       bufSize,
		flushInterval: flushInterval,
	}
}

func (s *deletionService) Submit(userID string, shortIDs []string) {
	if userID == "" || len(shortIDs) == 0 {
		return
	}
	go func(uid string, ids []string) {
		for _, id := range ids {
			select {
			case s.jobs <- deletionJob{userID: uid, shortID: id}:
			default:
				s.dropped.Add(1)
			}
		}
	}(userID, append([]string(nil), shortIDs...))
}

func (s *deletionService) Run(ctx context.Context) {
	buf := make([]deletionJob, 0, s.bufSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	flush := func() {
		if n := s.dropped.Swap(0); n > 0 {
			log.Printf("deletion buffer overflow: dropped %d updates", n)
		}
		if len(buf) == 0 {
			return
		}
		groups := make(map[string][]string, 8)
		for _, j := range buf {
			groups[j.userID] = append(groups[j.userID], j.shortID)
		}
		for uid, ids := range groups {
			if err := s.repo.DeleteUserURLs(uid, ids); err != nil {
				log.Printf("deletion flush failed user=%s n=%d: %v", uid, len(ids), err)
			}
		}
		buf = buf[:0]
	}

	for {
		select {
		case <-ctx.Done():
			for {
				select {
				case j := <-s.jobs:
					buf = append(buf, j)
					if len(buf) >= s.bufSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		case j := <-s.jobs:
			buf = append(buf, j)
			if len(buf) >= s.bufSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
