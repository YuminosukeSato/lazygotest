package tui

import (
	"sync"
)

// RingBuffer is a thread-safe circular buffer for storing log lines
type RingBuffer struct {
	maxSize int
	buffer  []string
	head    int // Index where next item will be written
	size    int // Current number of items in buffer
	mutex   sync.RWMutex
}

// NewRingBuffer creates a new ring buffer with specified max size
func NewRingBuffer(maxSize int) *RingBuffer {
	return &RingBuffer{
		maxSize: maxSize,
		buffer:  make([]string, maxSize),
		head:    0,
		size:    0,
	}
}

// Add adds a new line to the buffer, overwriting old lines if necessary
func (rb *RingBuffer) Add(line string) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.buffer[rb.head] = line
	rb.head = (rb.head + 1) % rb.maxSize
	
	if rb.size < rb.maxSize {
		rb.size++
	}
}

// AddMultiple adds multiple lines to the buffer
func (rb *RingBuffer) AddMultiple(lines []string) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	for _, line := range lines {
		rb.buffer[rb.head] = line
		rb.head = (rb.head + 1) % rb.maxSize
		
		if rb.size < rb.maxSize {
			rb.size++
		}
	}
}

// GetLines returns all lines in the buffer in chronological order
func (rb *RingBuffer) GetLines() []string {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	if rb.size == 0 {
		return []string{}
	}

	result := make([]string, rb.size)
	
	if rb.size < rb.maxSize {
		// Buffer not full yet, simply copy from start
		copy(result, rb.buffer[:rb.size])
	} else {
		// Buffer is full, need to reconstruct in correct order
		// Oldest items start at head position
		tail := rb.maxSize - rb.head
		copy(result[:tail], rb.buffer[rb.head:])
		copy(result[tail:], rb.buffer[:rb.head])
	}
	
	return result
}

// GetLinesRange returns a range of lines from start to end index
func (rb *RingBuffer) GetLinesRange(start, end int) []string {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	if rb.size == 0 || start >= rb.size || start < 0 {
		return []string{}
	}

	if end > rb.size {
		end = rb.size
	}

	lines := rb.getOrderedLines()
	return lines[start:end]
}

// Size returns the current number of lines in the buffer
func (rb *RingBuffer) Size() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	return rb.size
}

// Clear empties the buffer
func (rb *RingBuffer) Clear() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.head = 0
	rb.size = 0
	// Clear buffer to free memory
	for i := range rb.buffer {
		rb.buffer[i] = ""
	}
}

// getOrderedLines is a helper to get lines in chronological order (not thread-safe)
func (rb *RingBuffer) getOrderedLines() []string {
	if rb.size == 0 {
		return []string{}
	}

	result := make([]string, rb.size)
	
	if rb.size < rb.maxSize {
		copy(result, rb.buffer[:rb.size])
	} else {
		tail := rb.maxSize - rb.head
		copy(result[:tail], rb.buffer[rb.head:])
		copy(result[tail:], rb.buffer[:rb.head])
	}
	
	return result
}