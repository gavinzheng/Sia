package pool

import (
	"time"

	"github.com/NebulousLabs/Sia/build"
)

var (
	targetDuration   = build.Select(build.Var{
		Standard: 5.0,
		Dev:      5.0,
		Testing:  3.0,
	}).(float64) // targeted seconds between shares
	retargetDuration = build.Select(build.Var{
		Standard: 15.0,
		Dev:      15.0,
		Testing:  9.0,
	}).(float64) // how often do we consider changing difficulty
	variancePercent  = build.Select(build.Var{
		Standard: 15,
		Dev:      15,
		Testing:  15,
	}).(int)     // how much we let the share duration vary between retargetting
)

type Vardiff struct {
	tmax    float64
	tmin    float64
	bufSize uint64
}

func (s *Session) newVardiff() *Vardiff {
	variance := float64(targetDuration) * (float64(variancePercent) / 100.0)
	size := uint64(retargetDuration / targetDuration * 4)
	if size > numSharesToAverage {
		size = numSharesToAverage
	}
	v := &Vardiff{
		tmax:    targetDuration + variance,
		tmin:    targetDuration - variance,
		bufSize: numSharesToAverage,
	}

	s.lastVardiffRetarget = time.Now().Add(time.Duration(-retargetDuration / 2.0))
	s.lastVardiffTimestamp = time.Now()

	return v
}

func (s *Session) checkDiffOnNewShare() bool {

	s.lastVardiffTimestamp = time.Now()
	if time.Now().Sub(s.lastVardiffRetarget).Seconds() < retargetDuration {
		return false
	}
	if s.log != nil {
		s.log.Printf("Retargeted Duration: %f\n", time.Now().Sub(s.lastVardiffRetarget).Seconds())
	}
	s.lastVardiffRetarget = time.Now()

	unsubmitDuration, historyDuration := s.ShareDurationAverage()

	if unsubmitDuration > retargetDuration {
		s.SetCurrentDifficulty(s.CurrentDifficulty() / 2)
		if s.log != nil {
			s.log.Printf("UnsubmitDuration too long: %f, Set new difficulty to half: %v\n", unsubmitDuration, s.currentDifficulty)
		}
		return true
	}

	if historyDuration == 0 {
		if s.log != nil {
			s.log.Printf("No historyDuration yet\n")
		}
		return false
	}

	if historyDuration < s.vardiff.tmax && historyDuration > s.vardiff.tmin { // close enough
		if s.log != nil {
			s.log.Printf("HistoryDuration: %f is inside range\n", historyDuration)
		}
		return false
	}

	var deltaDiff float64
	deltaDiff = float64(targetDuration) / float64(historyDuration)
	deltaDiff = 1 - (1 - deltaDiff)/2
	if deltaDiff > 2.0 {
		deltaDiff = 2.0
	}
	if deltaDiff < 0.5 {
		deltaDiff = 0.5
	}

	if s.log != nil {
		s.log.Printf("HistoryDuration: %f Delta %f\n", historyDuration, deltaDiff)
	}

	if s.log != nil {
		s.log.Printf("Old difficulty was %v\n", s.currentDifficulty)
	}
	s.SetCurrentDifficulty(s.CurrentDifficulty() * deltaDiff)
	if s.log != nil {
		s.log.Printf("Set new difficulty to %v\n", s.currentDifficulty)
	}
	err := s.CurrentWorker.updateWorkerRecord()
	if s.log != nil && err != nil {
		s.log.Printf("Error updating worker record: %s\n", err)
	}
	return true
}
