package mogushan

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
)

func areAllAlive(arr []ns4.Obj) bool {
	for _, u := range arr {
		if u.CurrentHealth() > 0 {
			return true
		}
	}
	return false
}
