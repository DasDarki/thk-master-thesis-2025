package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

var runManager = newRunManager()

type RunManager struct {
	runs map[uuid.UUID]*TestRun
	mu   sync.RWMutex
}

func newRunManager() *RunManager {
	return &RunManager{
		runs: map[uuid.UUID]*TestRun{},
	}
}

func (rm *RunManager) beginRun(protocol Protocol, env Enviroment, timeSlot TimeSlot, clientID, parallelClients int) *TestRun {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	run := &TestRun{
		ID:              uuid.New(),
		Protocol:        protocol,
		Enviroment:      env,
		TimeSlot:        timeSlot,
		ClientID:        clientID,
		ParallelClients: parallelClients,
		TestBegin:       time.Now(),
	}

	rm.runs[run.ID] = run

	return run
}

func (rm *RunManager) endRun(runID uuid.UUID, data TestRunData) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	run := rm.runs[runID]
	run.TestEnd = time.Now()
	run.Data = data

	delete(rm.runs, runID)

	if err := saveToCsv(run); err != nil {
		if err := saveToEmergencyCsv(run); err != nil {
			return err
		}
	}

	return nil
}
