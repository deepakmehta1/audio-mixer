package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	regularQueue  []string // loaded from files/songs.json and maintained circularly
	priorityQueue []string // maximum 2 songs only
	queueMutex    sync.Mutex
)

// LoadRegularQueue loads the regular queue from the specified JSON file.
func LoadRegularQueue(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var songs []string
	if err := json.Unmarshal(data, &songs); err != nil {
		return err
	}
	queueMutex.Lock()
	regularQueue = songs
	queueMutex.Unlock()
	log.Printf("Loaded %d songs into regular queue from %s", len(songs), path)
	return nil
}

// NextSong returns the next song to play.
// It checks the priority queue first; if empty, it pops from the regular queue in a circular fashion.
func NextSong() string {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	if len(priorityQueue) > 0 {
		song := priorityQueue[0]
		priorityQueue = priorityQueue[1:]
		return song
	}
	if len(regularQueue) > 0 {
		song := regularQueue[0]
		regularQueue = append(regularQueue[1:], song)
		return song
	}
	return ""
}

// AddPrioritySong adds a song to the priority queue.
func AddPrioritySong(path string) error {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	if len(priorityQueue) >= 20 {
		return fmt.Errorf("priority queue is full (max 20 songs allowed)")
	}
	priorityQueue = append(priorityQueue, path)
	log.Printf("Added priority song to queue: %s", path)
	return nil
}

// AddRegularSong adds a song to the regular queue.
func AddRegularSong(path string) {
	queueMutex.Lock()
	regularQueue = append(regularQueue, path)
	queueMutex.Unlock()
	log.Printf("Added regular song to queue: %s", path)
}

// GetPriorityQueue returns a copy of the current priority queue.
func GetPriorityQueue() []string {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	copyQ := make([]string, len(priorityQueue))
	copy(copyQ, priorityQueue)
	return copyQ
}
